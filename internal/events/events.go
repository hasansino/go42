package events

import (
	"context"
)

// Publisher publishes event to the topic in async fashion.
type Publisher interface {
	Publish(topic string, event []byte) error
}

// Subscriber subscribes a handler for given topic in async fashion.
// Passed context control underlying goroutine and terminates in upon canceling.
// Handler should return error for Nack or nil to Ack.
type Subscriber interface {
	Subscribe(
		ctx context.Context, topic string,
		handler func(ctx context.Context, event []byte) error,
	) error
}

type Eventer interface {
	Publisher
	Subscriber
	Shutdown(ctx context.Context) error
}

// ---

// NoopEngine is a no-op implementation of Eventer.
type NoopEngine struct{}

func NewNoop() *NoopEngine {
	return &NoopEngine{}
}

func (e *NoopEngine) Publish(_ string, _ []byte) error {
	return nil
}

func (e *NoopEngine) Subscribe(
	_ context.Context, _ string,
	_ func(_ context.Context, _ []byte) error,
) error {
	return nil
}

func (e *NoopEngine) Shutdown(_ context.Context) error {
	return nil
}
