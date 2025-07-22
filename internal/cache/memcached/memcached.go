package memcached

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/bradfitz/gomemcache/memcache"
)

type Wrapper struct {
	client *memcache.Client
}

func Open(ctx context.Context, hosts []string, opts ...Option) (*Wrapper, error) {
	client, err := retry.DoWithData[*memcache.Client](func() (*memcache.Client, error) {
		client := memcache.New(hosts...)
		for _, opt := range opts {
			opt(client)
		}
		return client, client.Ping()
	},
		retry.Context(ctx),
		retry.Attempts(10),
		retry.Delay(2*time.Second),
		retry.MaxDelay(2*time.Second),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			slog.Default().WarnContext(
				ctx,
				"cache connection attempt failed, retrying...",
				slog.String("component", "memcached"),
				slog.Any("attempt", n+1),
				slog.String("error", err.Error()),
			)
		}),
	)
	if err != nil {
		return nil, err
	}
	return &Wrapper{client: client}, client.Ping()
}

func (w *Wrapper) Shutdown(ctx context.Context) error {
	done := make(chan error)
	go func() {
		done <- w.client.Close()
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}

func (w *Wrapper) Get(_ context.Context, key string) (string, error) {
	item, err := w.client.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (w *Wrapper) Set(_ context.Context, key string, value string) error {
	return w.client.Set(&memcache.Item{Key: key, Value: []byte(value)})
}

func (w *Wrapper) SetTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	return w.client.Set(&memcache.Item{Key: key, Value: []byte(value), Expiration: int32(ttl.Seconds())})
}

func (w *Wrapper) Invalidate(_ context.Context, key string) error {
	return w.client.Delete(key)
}
