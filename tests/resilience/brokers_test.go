//go:build resilience

package resilience

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/hasansino/go42/internal/events"
)

const (
	brokerTestTimeout      = 30 * time.Second
	brokerLatencyToxicName = "broker_response_latency"
)

func TestBrokerSubscribersRecoverAfterNetworkInterruption(t *testing.T) {
	for _, factory := range brokerDependencyFactories() {
		if factory.raceSensitive {
			continue
		}
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			testBrokerSubscriberRecovery(t, factory)
		})
	}
}

func TestRabbitMQSubscriberRecoversAfterNetworkInterruption(t *testing.T) {
	if raceDetectorEnabled {
		t.Skip("watermill-amqp reconnect has a known upstream data race")
	}
	testBrokerSubscriberRecovery(t, rabbitMQFactory())
}

func TestBrokersRecoverFromResponseLatency(t *testing.T) {
	for _, factory := range brokerDependencyFactories() {
		if factory.raceSensitive {
			continue
		}
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			testBrokerLatencyRecovery(t, factory)
		})
	}
}

func TestRabbitMQRecoversFromResponseLatency(t *testing.T) {
	if raceDetectorEnabled {
		t.Skip("watermill-amqp reconnect has a known upstream data race")
	}
	testRabbitMQLatencyRecovery(t, rabbitMQFactory())
}

func TestBrokersShutdownWhileDisconnected(t *testing.T) {
	for _, factory := range brokerDependencyFactories() {
		if factory.raceSensitive {
			continue
		}
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			testBrokerShutdownWhileDisconnected(t, factory)
		})
	}
}

func TestRabbitMQShutdownWhileDisconnected(t *testing.T) {
	if raceDetectorEnabled {
		t.Skip("watermill-amqp reconnect has a known upstream data race")
	}
	testBrokerShutdownWhileDisconnected(t, rabbitMQFactory())
}

func testBrokerSubscriberRecovery(t *testing.T, factory brokerFactory) {
	t.Helper()
	resetProxy(t, factory.proxy)
	ctx, cancel := context.WithTimeout(t.Context(), brokerTestTimeout)
	defer cancel()
	backend, err := factory.open(ctx, startupTestTimeout)
	if err != nil {
		t.Fatalf("initialize broker: %v", err)
	}
	defer shutdownBackend(t, backend)

	topic := uniqueTopic("delivery")
	assertEventuallySucceeds(t, "create broker topic", func() error {
		return publishMessage(ctx, backend, topic)
	})
	messages, err := backend.Subscriber().Subscribe(ctx, topic)
	if err != nil {
		t.Fatalf("subscribe before outage: %v", err)
	}
	assertMessageRoundTrip(t, ctx, backend, messages, topic)

	setProxyEnabled(t, factory.proxy.Name, false)
	assertEventuallyFails(t, "publish while broker is unavailable", func() error {
		return publishMessage(ctx, backend, topic)
	})
	setProxyEnabled(t, factory.proxy.Name, true)
	assertMessageRoundTrip(t, ctx, backend, messages, topic)

	errs := publishConcurrently(ctx, backend, uniqueTopic("concurrent"), concurrentOperationCount)
	for index, err := range errs {
		if err != nil {
			t.Errorf("concurrent publish %d failed after recovery: %v", index, err)
		}
	}
}

func testBrokerLatencyRecovery(t *testing.T, factory brokerFactory) {
	t.Helper()
	resetProxy(t, factory.proxy)
	ctx, cancel := context.WithTimeout(t.Context(), brokerTestTimeout)
	defer cancel()
	backend, err := factory.open(ctx, startupTestTimeout)
	if err != nil {
		t.Fatalf("initialize broker: %v", err)
	}
	defer shutdownBackend(t, backend)

	topic := uniqueTopic("latency")
	if err := publishMessage(ctx, backend, topic); err != nil {
		t.Fatalf("publish before latency injection: %v", err)
	}
	addToxic(t, factory.proxy.Name, toxicConfig{
		Name:     brokerLatencyToxicName,
		Type:     "latency",
		Stream:   "downstream",
		Toxicity: 1,
		Attributes: map[string]any{
			"latency": 6000,
			"jitter":  0,
		},
	})
	assertEventuallyFails(t, "publish while broker responses exceed timeout", func() error {
		operationCtx, operationCancel := context.WithTimeout(ctx, time.Second)
		defer operationCancel()
		return publishMessage(operationCtx, backend, topic)
	})

	removeToxic(t, factory.proxy.Name, brokerLatencyToxicName)
	assertEventuallySucceeds(t, "publish after broker latency recovers", func() error {
		return publishMessage(ctx, backend, topic)
	})
}

func testRabbitMQLatencyRecovery(t *testing.T, factory brokerFactory) {
	t.Helper()
	resetProxy(t, factory.proxy)
	ctx, cancel := context.WithTimeout(t.Context(), brokerTestTimeout)
	defer cancel()
	backend, err := factory.open(ctx, startupTestTimeout)
	if err != nil {
		t.Fatalf("initialize broker: %v", err)
	}
	defer shutdownBackend(t, backend)

	topic := uniqueTopic("latency")
	addToxic(t, factory.proxy.Name, toxicConfig{
		Name:     brokerLatencyToxicName,
		Type:     "latency",
		Stream:   "downstream",
		Toxicity: 1,
		Attributes: map[string]any{
			"latency": 500,
			"jitter":  0,
		},
	})
	operationCtx, operationCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	started := time.Now()
	_ = publishMessage(operationCtx, backend, topic)
	operationCancel()
	if elapsed := time.Since(started); elapsed < 400*time.Millisecond {
		t.Errorf("RabbitMQ response latency was not applied: publish returned after %s", elapsed)
	}

	removeToxic(t, factory.proxy.Name, brokerLatencyToxicName)
	assertEventuallySucceeds(t, "publish after RabbitMQ latency recovers", func() error {
		return publishMessage(ctx, backend, topic)
	})
}

func testBrokerShutdownWhileDisconnected(t *testing.T, factory brokerFactory) {
	t.Helper()
	resetProxy(t, factory.proxy)
	backend, err := factory.open(t.Context(), startupTestTimeout)
	if err != nil {
		t.Fatalf("initialize broker: %v", err)
	}
	topic := uniqueTopic("shutdown")
	setProxyEnabled(t, factory.proxy.Name, false)
	assertEventuallyFails(t, "observe broker outage before shutdown", func() error {
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		return publishMessage(ctx, backend, topic)
	})

	shutdownCtx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	started := time.Now()
	err = backend.Shutdown(shutdownCtx)
	if elapsed := time.Since(started); elapsed > 4*time.Second {
		t.Errorf("shutdown exceeded its deadline: elapsed %s, error %v", elapsed, err)
	}
}

func assertMessageRoundTrip(
	t *testing.T,
	ctx context.Context,
	backend events.Backend,
	messages <-chan *message.Message,
	topic string,
) {
	t.Helper()
	wantID := watermill.NewUUID()
	msg := message.NewMessage(wantID, []byte("resilience check"))
	msg.SetContext(ctx)
	assertEventuallySucceeds(t, "publish message for round trip", func() error {
		return backend.Publisher().Publish(topic, msg)
	})

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("receive message %q: %v", wantID, ctx.Err())
		case received, open := <-messages:
			if !open {
				t.Fatalf("subscriber closed before receiving message %q", wantID)
			}
			if received.UUID == wantID {
				received.Ack()
				return
			}
			received.Ack()
		}
	}
}

func publishMessage(ctx context.Context, backend events.Backend, topic string) error {
	msg := message.NewMessage(watermill.NewUUID(), []byte("resilience check"))
	msg.SetContext(ctx)
	return backend.Publisher().Publish(topic, msg)
}

func publishConcurrently(
	ctx context.Context,
	backend events.Backend,
	topic string,
	count int,
) []error {
	errs := make([]error, count)
	var waitGroup sync.WaitGroup
	waitGroup.Add(count)
	for index := range count {
		go func() {
			defer waitGroup.Done()
			operationCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			errs[index] = publishMessage(operationCtx, backend, topic)
		}()
	}
	waitGroup.Wait()
	return errs
}

func shutdownBackend(t *testing.T, backend events.Backend) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := backend.Shutdown(ctx); err != nil {
		t.Errorf("shutdown broker: %v", err)
	}
}

func rabbitMQFactory() brokerFactory {
	for _, factory := range brokerDependencyFactories() {
		if factory.name == "RabbitMQ" {
			return factory
		}
	}
	panic(fmt.Errorf("RabbitMQ resilience factory is missing"))
}
