package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Wrapper struct {
	client *redis.Client
}

func New(host string, db int, opts ...Option) (*Wrapper, error) {
	w := new(Wrapper)
	cfg := &redis.Options{
		Addr: host,
		DB:   db,
	}
	for _, opt := range opts {
		opt(w, cfg)
	}
	rdb := redis.NewClient(cfg)
	status := rdb.Ping(context.Background())
	if status.Err() != nil {
		return nil, status.Err()
	}
	w.client = rdb
	return w, nil
}

func (w *Wrapper) Shutdown(ctx context.Context) error {
	doneChan := make(chan error)
	go func() {
		doneChan <- w.client.Shutdown(ctx).Err()
	}()
	select {
	case <-ctx.Done():
		return errors.New("timeout")
	case err := <-doneChan:
		return err
	}
}

func (w *Wrapper) Get(ctx context.Context, key string) (string, error) {
	cmd := w.client.Get(ctx, key)
	if cmd.Err() != nil {
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
