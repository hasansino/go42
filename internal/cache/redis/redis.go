package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/redis/go-redis/v9"
)

type Wrapper struct {
	client *redis.Client
}

// allowRateLimitScript implements GCRA atomically inside Redis.
var allowRateLimitScript = redis.NewScript(`
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local ttl_ms = tonumber(ARGV[3])
local interval = math.max(1, math.floor(1000000 / rate))
local now_parts = redis.call('TIME')
local now = (tonumber(now_parts[1]) * 1000000) + tonumber(now_parts[2])
local tat = tonumber(redis.call('GET', KEYS[1]))

if not tat or tat < now then
	tat = now
end

local allow_at = tat - ((burst - 1) * interval)
if now < allow_at then
	redis.call('PEXPIRE', KEYS[1], ttl_ms)
	return 0
end

redis.call('SET', KEYS[1], string.format('%.0f', tat + interval), 'PX', ttl_ms)
return 1
`)

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

func (w *Wrapper) AllowRateLimit(
	ctx context.Context,
	key string,
	rate int,
	burst int,
	ttl time.Duration,
) (bool, error) {
	if rate <= 0 {
		return false, fmt.Errorf("rate must be positive")
	}
	if burst <= 0 {
		return false, fmt.Errorf("burst must be positive")
	}
	if ttl <= 0 {
		return false, fmt.Errorf("ttl must be positive")
	}

	ttlMilliseconds := ttl.Milliseconds()
	if ttlMilliseconds < 1 {
		ttlMilliseconds = 1
	}

	allowed, err := allowRateLimitScript.Run(
		ctx,
		w.client,
		[]string{key},
		rate,
		burst,
		ttlMilliseconds,
	).Int()
	if err != nil {
		return false, err
	}
	return allowed == 1, nil
}

func (w *Wrapper) Invalidate(ctx context.Context, key string) error {
	cmd := w.client.Del(ctx, key)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}
