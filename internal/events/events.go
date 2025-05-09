package events

import "context"

type Eventer interface {
	// Publish publishes event to the topic. It should be non-blocking.
	Publish(topic string, event []byte) error
	// Subscribe starts goroutine to listen to the topic.
	// Context cancel should be used to cancel the subscription.
	// Handler should return error to avoid Ack of the message.
	Subscribe(
		ctx context.Context, topic string,
		handler func(ctx context.Context, event []byte) error,
	) error
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
