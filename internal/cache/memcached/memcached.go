package memcached

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
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
	done := make(chan error, 1)
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

func (w *Wrapper) AllowRateLimit(
	ctx context.Context,
	key string,
	rate int,
	burst int,
	ttl time.Duration,
) (bool, error) {
	// Memcached has no server-side scripting, so GCRA state is updated with CAS.
	if rate <= 0 {
		return false, fmt.Errorf("rate must be positive")
	}
	if burst <= 0 {
		return false, fmt.Errorf("burst must be positive")
	}
	if ttl <= 0 {
		return false, fmt.Errorf("ttl must be positive")
	}

	intervalMicros := (time.Second / time.Duration(rate)).Microseconds()
	if intervalMicros < 1 {
		intervalMicros = 1
	}
	burstTolerance := int64(burst-1) * intervalMicros

	const maxCASAttempts = 16
	for range maxCASAttempts {
		if err := ctx.Err(); err != nil {
			return false, err
		}

		now := time.Now().UnixMicro()
		item, err := w.client.Get(key)
		if errors.Is(err, memcache.ErrCacheMiss) {
			item = newItem(key, strconv.FormatInt(now+intervalMicros, 10), ttl)
			err = w.client.Add(item)
			if err == nil {
				return true, nil
			}
			if errors.Is(err, memcache.ErrNotStored) {
				continue
			}
			return false, err
		}
		if err != nil {
			return false, err
		}

		tat, err := strconv.ParseInt(string(item.Value), 10, 64)
		if err != nil {
			return false, fmt.Errorf("decode rate limiter state: %w", err)
		}
		if tat < now {
			tat = now
		}
		if now < tat-burstTolerance {
			return false, nil
		}

		item.Value = []byte(strconv.FormatInt(tat+intervalMicros, 10))
		item.Expiration = expirationSeconds(ttl)
		err = w.client.CompareAndSwap(item)
		if err == nil {
			return true, nil
		}
		if errors.Is(err, memcache.ErrCASConflict) || errors.Is(err, memcache.ErrNotStored) {
			continue
		}
		return false, err
	}

	return false, fmt.Errorf("rate limiter state remained contended after %d attempts", maxCASAttempts)
}

func (w *Wrapper) Invalidate(_ context.Context, key string) error {
	err := w.client.Delete(key)
	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil
	}
	return err
}

func newItem(key string, value string, ttl time.Duration) *memcache.Item {
	item := &memcache.Item{Key: key, Value: []byte(value)}
	if ttl > 0 {
		item.Expiration = expirationSeconds(ttl)
	}
	return item
}

func expirationSeconds(ttl time.Duration) int32 {
	if ttl <= 0 {
		return 0
	}

	seconds := ttl / time.Second
	if ttl%time.Second != 0 {
		seconds++
	}
	if seconds > time.Duration(1<<31-1) {
		return 1<<31 - 1
	}

	return int32(seconds)
}
