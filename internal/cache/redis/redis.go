package redis

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/redis/go-redis/v9"
)

type Wrapper struct {
	client *redis.Client
}

func Open(ctx context.Context, host string, db int, opts ...Option) (*Wrapper, error) {
	w := new(Wrapper)

	cfg := &redis.Options{
		Addr: host,
		DB:   db,
	}
	for _, opt := range opts {
		opt(w, cfg)
	}

	rdb, err := retry.DoWithData[*redis.Client](func() (*redis.Client, error) {
		rdb := redis.NewClient(cfg)
		status := rdb.Ping(context.Background())
		if status.Err() != nil {
			return nil, status.Err()
		}
		return rdb, nil
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
				slog.Int("attempt", int(n+1)),
				slog.String("error", err.Error()),
			)
		}),
	)
	if err != nil {
		return nil, err
	}

	w.client = rdb
	return w, nil
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

func (w *Wrapper) Get(ctx context.Context, key string) (string, error) {
	cmd := w.client.Get(ctx, key)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return "", nil
		}
		return "", cmd.Err()
	}
	return cmd.Val(), nil
}

func (w *Wrapper) Set(ctx context.Context, key string, value string) error {
	return w.SetTTL(ctx, key, value, 0)
}

func (w *Wrapper) SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error {
	cmd := w.client.Set(ctx, key, value, ttl)
	return cmd.Err()
}
