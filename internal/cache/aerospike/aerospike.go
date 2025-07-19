package aerospike

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	aslib "github.com/aerospike/aerospike-client-go/v8"
	asTypes "github.com/aerospike/aerospike-client-go/v8/types"
	"github.com/avast/retry-go/v4"
)

const (
	defaultSet = "default"
)

type Wrapper struct {
	client    *aslib.Client
	namespace string
}

func Open(ctx context.Context, hosts []string, namespace string, opts ...Option) (*Wrapper, error) {
	w := &Wrapper{
		namespace: namespace,
	}

	policy := aslib.NewClientPolicy()

	for _, opt := range opts {
		opt(policy)
	}

	parsedHosts, err := parseHosts(hosts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hosts: %w", err)
	}

	client, err := retry.DoWithData[*aslib.Client](func() (*aslib.Client, error) {
		client, err := aslib.NewClientWithPolicyAndHost(policy, parsedHosts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create aerospike client: %w", err)
		}
		_, err = client.WarmUp(len(parsedHosts))
		if err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to warm up aerospike client: %w", err)
		}
		return client, nil
	},
		retry.Context(ctx),
		retry.Attempts(5),
		retry.Delay(2*time.Second),
		retry.MaxDelay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			slog.Default().WarnContext(
				ctx,
				"cache connection attempt failed, retrying...",
				slog.String("component", "redis"),
				slog.Any("attempt", n+1),
				slog.String("error", err.Error()),
			)
		}),
	)
	if err != nil {
		return nil, err
	}

	w.client = client

	return w, nil
}

func (w *Wrapper) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		w.client.Close()
		done <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case <-done:
		return nil
	}
}

func (w *Wrapper) Get(_ context.Context, key string) (string, error) {
	asKey, err := aslib.NewKey(w.namespace, defaultSet, key)
	if err != nil {
		return "", fmt.Errorf("failed to create aerospike key: %w", err)
	}
	r, err := w.client.Get(w.client.GetDefaultPolicy(), asKey, key)
	if err != nil {
		if err.Matches(asTypes.KEY_NOT_FOUND_ERROR) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get aerospike value: %w", err)
	}
	binValue := r.Bins[key]
	if binValue == nil {
		return "", nil
	}
	return binValue.(string), nil
}

func (w *Wrapper) Set(_ context.Context, key string, value string) error {
	asKey, err := aslib.NewKey(w.namespace, defaultSet, key)
	if err != nil {
		return fmt.Errorf("failed to create aerospike key: %w", err)
	}
	err = w.client.Put(w.client.GetDefaultWritePolicy(), asKey, map[string]interface{}{key: value})
	if err != nil {
		return fmt.Errorf("failed to set aerospike value: %w", err)
	}
	return nil
}

func (w *Wrapper) SetTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	asKey, err := aslib.NewKey(w.namespace, defaultSet, key)
	if err != nil {
		return fmt.Errorf("failed to create aerospike key: %w", err)
	}
	policy := w.client.GetDefaultWritePolicy()
	policy.Expiration = uint32(ttl.Seconds())
	err = w.client.Put(policy, asKey, map[string]interface{}{key: value})
	if err != nil {
		return fmt.Errorf("failed to set aerospike value: %w", err)
	}
	return nil
}

func parseHosts(hosts []string) ([]*aslib.Host, error) {
	var parsedHosts []*aslib.Host
	for _, host := range hosts {
		parts := strings.Split(host, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid host format: %s", host)
		}
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port format: %s", parts[1])
		}
		parsedHosts = append(parsedHosts, aslib.NewHost(parts[0], port))
	}
	return parsedHosts, nil
}
