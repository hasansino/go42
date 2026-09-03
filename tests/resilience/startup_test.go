//go:build resilience

package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

const (
	startupRecoveryDelay = 200 * time.Millisecond
	startupTestTimeout   = 10 * time.Second
	startupRetryDeadline = 400 * time.Millisecond
)

type openResult struct {
	client *resilienceClient
	err    error
}

func TestDependenciesRetryStartupUntilAvailable(t *testing.T) {
	for _, factory := range allDependencyFactories() {
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			resetProxy(t, factory.proxy)
			setProxyEnabled(t, factory.proxy.Name, false)

			ctx, cancel := context.WithTimeout(t.Context(), startupTestTimeout)
			defer cancel()
			result := make(chan openResult, 1)
			go func() {
				client, err := factory.open(ctx, startupTestTimeout)
				result <- openResult{client: client, err: err}
			}()

			select {
			case result := <-result:
				if result.client != nil {
					shutdownClient(t, result.client)
				}
				t.Fatalf("dependency initialization returned before recovery: %v", result.err)
			case <-time.After(startupRecoveryDelay):
			}

			setProxyEnabled(t, factory.proxy.Name, true)
			select {
			case result := <-result:
				if result.err != nil {
					t.Fatalf("dependency did not initialize after recovery: %v", result.err)
				}
				defer shutdownClient(t, result.client)
				assertOperationSucceeds(t, result.client.operation)
			case <-ctx.Done():
				t.Fatalf("dependency did not initialize after recovery: %v", ctx.Err())
			}
		})
	}
}

func TestDependenciesRespectStartupRetryDeadline(t *testing.T) {
	for _, factory := range allDependencyFactories() {
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			resetProxy(t, factory.proxy)
			setProxyEnabled(t, factory.proxy.Name, false)

			started := time.Now()
			client, err := factory.open(t.Context(), startupRetryDeadline)
			if client != nil {
				shutdownClient(t, client)
				t.Fatal("dependency initialized while its proxy was disabled")
			}
			if err == nil {
				t.Fatal("dependency initialization returned no error after retry deadline")
			}
			if elapsed := time.Since(started); elapsed < startupRetryDeadline/2 {
				t.Errorf("dependency stopped retrying too early after %s", elapsed)
			}
		})
	}
}

func TestDependenciesRespectStartupContextCancellation(t *testing.T) {
	for _, factory := range allDependencyFactories() {
		factory := factory
		t.Run(factory.name, func(t *testing.T) {
			resetProxy(t, factory.proxy)
			setProxyEnabled(t, factory.proxy.Name, false)

			ctx, cancel := context.WithCancel(t.Context())
			time.AfterFunc(startupRecoveryDelay, cancel)
			started := time.Now()
			client, err := factory.open(ctx, startupTestTimeout)
			if client != nil {
				shutdownClient(t, client)
				t.Fatal("dependency initialized while its proxy was disabled")
			}
			if !errors.Is(err, context.Canceled) {
				t.Errorf("dependency initialization error = %v, want context cancellation", err)
			}
			if elapsed := time.Since(started); elapsed > 3*time.Second {
				t.Errorf("dependency took %s to honor context cancellation", elapsed)
			}
		})
	}
}
