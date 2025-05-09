package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"

	"github.com/hasansino/goapp/internal/example"
)

type Watermill interface {
	Publisher() message.Publisher
	Subscriber() message.Subscriber
}

type Provider struct {
	logger  *slog.Logger
	service *example.Service
	router  *message.Router
}

func New(
	service *example.Service, eventer Watermill, opts ...Option,
) (*Provider, error) {
	var (
		provider = new(Provider)
	)

	for _, opt := range opts {
		opt(provider)
	}

	if provider.logger == nil {
		provider.logger = slog.New(slog.DiscardHandler)
	}

	provider.service = service

	router, err := message.NewRouter(
		message.RouterConfig{
			CloseTimeout: 5 * time.Second,
		},
		watermill.NewSlogLogger(provider.logger),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create router: %w", err)
	}

	router.AddHandler(
		"example-watermill-provider",
		"example-topic",
		eventer.Subscriber(),
		"example-topic",
		eventer.Publisher(),
		provider.Handle,
	)

	router.AddMiddleware(
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			Logger:          watermill.NewSlogLogger(provider.logger),
		}.Middleware,
		middleware.Recoverer,
	)

	provider.router = router

	return provider, nil
}

func (p *Provider) Run(ctx context.Context) error {
	return p.router.Run(ctx)
}

func (p *Provider) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		done <- p.router.Close()
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}

func (p *Provider) Handle(msg *message.Message) ([]*message.Message, error) {
	return nil, p.service.ProcessEvent(msg.Context(), msg.Payload)
}
