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
		retry.Attempts(10),
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

func (w *Wrapper) Get(ctx context.Context, key string) (string, bool, error) {
	cmd := w.client.Get(ctx, key)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return "", false, nil
		}
		return "", false, cmd.Err()
	}
	return cmd.Val(), true, nil
}

func (w *Wrapper) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	cmd := w.client.Set(ctx, key, value, ttl)
	return cmd.Err()
}

func (w *Wrapper) SetIfAbsent(
	ctx context.Context,
	key string,
	value string,
	ttl time.Duration,
) (bool, error) {
	err := w.client.SetArgs(ctx, key, value, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Err()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (w *Wrapper) Invalidate(ctx context.Context, key string) error {
	cmd := w.client.Del(ctx, key)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}
