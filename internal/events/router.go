package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	"github.com/go42-dev/go42/internal/metrics"
)

type permanentError struct {
	err error
}

func (e *permanentError) Error() string {
	return e.err.Error()
}

func (e *permanentError) Unwrap() error {
	return e.err
}

func Permanent(err error) error {
	if err == nil {
		return nil
	}

	var permanent *permanentError
	if errors.As(err, &permanent) {
		return err
	}

	return &permanentError{err: err}
}

// ---

type Router struct {
	backend         Backend
	router          *message.Router
	policy          DeliveryPolicy
	logger          *slog.Logger
	watermillLogger watermill.LoggerAdapter
	handlerID       atomic.Uint64
}

type DeliveryPolicy struct {
	MaxRetries            int
	InitialBackoff        time.Duration
	MaxBackoff            time.Duration
	DeadLetterTopicSuffix string
	CloseTimeout          time.Duration
}

func NewRouter(backend Backend, policy DeliveryPolicy, opts ...Option) (*Router, error) {
	if backend == nil {
		return nil, errors.New("event backend is required")
	}

	r := &Router{
		backend: backend,
		policy:  policy,
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.logger == nil {
		r.logger = slog.New(slog.DiscardHandler)
	}

	watermillLogger := watermill.NewSlogLogger(r.logger)
	watermillRouter, err := message.NewRouter(message.RouterConfig{
		CloseTimeout: policy.CloseTimeout,
	}, watermillLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Watermill router: %w", err)
	}

	r.router = watermillRouter
	r.watermillLogger = watermillLogger

	return r, nil
}

func (r *Router) Publish(topic string, event []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), event)
	err := r.backend.Publisher().Publish(topic, msg)
	result := "success"
	if err != nil {
		result = "error"
	}
	metrics.Counter("application_event_publish_total", map[string]any{
		"result": result,
		"topic":  topic,
	}).Inc()
	return err
}

func (r *Router) Subscribe(
	topic string,
	handler func(ctx context.Context, event []byte) error,
) error {
	deadLetterTopic := topic + r.policy.DeadLetterTopicSuffix

	if initializer, ok := r.backend.(TopicInitializer); ok {
		if err := initializer.InitializeTopic(deadLetterTopic); err != nil {
			return fmt.Errorf(
				"failed to initialize dead-letter topic %q: %w",
				deadLetterTopic,
				err,
			)
		}
	}

	poisonQueue, err := middleware.PoisonQueue(
		&deadLetterPublisher{
			publisher: r.backend.Publisher(),
			logger:    r.logger,
		},
		deadLetterTopic,
	)
	if err != nil {
		return fmt.Errorf("failed to configure dead-letter topic %q: %w", deadLetterTopic, err)
	}

	retry := middleware.Retry{
		MaxRetries:          r.policy.MaxRetries,
		InitialInterval:     r.policy.InitialBackoff,
		MaxInterval:         r.policy.MaxBackoff,
		Multiplier:          2,
		RandomizationFactor: 0.2,
		ResetContextOnRetry: true,
		Logger:              r.watermillLogger,
		ShouldRetry: func(params middleware.RetryParams) bool {
			var permanent *permanentError
			return !errors.As(params.Err, &permanent)
		},
	}

	watermillHandler := r.router.AddConsumerHandler(
		fmt.Sprintf("%s_%d", topic, r.handlerID.Add(1)),
		topic,
		r.backend.Subscriber(),
		func(msg *message.Message) error {
			return handler(msg.Context(), msg.Payload)
		},
	)

	// `PoisonQueue` must wrap `Retry` so it sees only the final error.
	watermillHandler.AddMiddleware(poisonQueue, retryWithMetrics(topic, retry))

	return nil
}

func retryWithMetrics(topic string, retry middleware.Retry) message.HandlerMiddleware {
	retryCounter := metrics.Counter("application_event_consumer_retries_total", map[string]any{
		"topic": topic,
	})

	return func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			attempt := 0
			retryHandler := retry.Middleware(func(msg *message.Message) ([]*message.Message, error) {
				if attempt > 0 {
					retryCounter.Inc()
				}
				attempt++
				return next(msg)
			})
			return retryHandler(msg)
		}
	}
}

func (r *Router) Start(ctx context.Context) error {
	runErr := make(chan error, 1)
	go func() {
		runErr <- r.router.Run(ctx)
	}()

	select {
	case <-r.router.Running():
	case err := <-runErr:
		if err == nil {
			return errors.New("event router stopped during startup")
		}
		return fmt.Errorf("failed to start event router: %w", err)
	}

	go func() {
		err := <-runErr
		// if ctx is canceled, we don't want to log an error because it's expected behavior
		if ctx.Err() != nil {
			return
		}
		metrics.Counter("application_event_router_stops_total", map[string]any{
			"reason": "unexpected",
		}).Inc()
		// otherwise this is unexpected and we should log it
		if err != nil {
			r.logger.ErrorContext(ctx, "event router stopped", slog.Any("error", err))
			return
		}
		// should not happen, but log it just in case
		r.logger.ErrorContext(ctx, "event router stopped unexpectedly without error")
	}()

	return nil
}

func (r *Router) Shutdown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		routerErr := r.router.Close()
		backendErr := r.backend.Shutdown(ctx)
		done <- errors.Join(routerErr, backendErr)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
