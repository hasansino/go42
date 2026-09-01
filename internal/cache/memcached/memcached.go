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

func (w *Wrapper) Get(_ context.Context, key string) (string, bool, error) {
	item, err := w.client.Get(key)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return "", false, nil
		}
		return "", false, err
	}
	return string(item.Value), true, nil
}

func (w *Wrapper) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	return w.client.Set(newItem(key, value, ttl))
}

func (w *Wrapper) SetIfAbsent(
	_ context.Context,
	key string,
	value string,
	ttl time.Duration,
) (bool, error) {
	err := w.client.Add(newItem(key, value, ttl))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, memcache.ErrNotStored) {
		return false, nil
	}
	return false, err
}

func (w *Wrapper) Invalidate(_ context.Context, key string) error {
	return w.client.Delete(key)
}

func newItem(key string, value string, ttl time.Duration) *memcache.Item {
	item := &memcache.Item{Key: key, Value: []byte(value)}
	if ttl > 0 {
		item.Expiration = int32(ttl.Seconds())
	}
	return item
}
