package miniredis

import (
	"context"
	"time"

	"github.com/alicebob/miniredis/v2"
)

type Wrapper struct {
	client *miniredis.Miniredis
}

func New() *Wrapper {
	return &Wrapper{
		client: miniredis.NewMiniRedis(),
	}
}

func (w *Wrapper) Close() error {
	w.client.Close()
	return nil
}

func (w *Wrapper) Get(_ context.Context, key string) (string, error) {
	return w.client.Get(key)
}

func (w *Wrapper) Set(_ context.Context, key string, value string) error {
	return w.client.Set(key, value)
}

func (w *Wrapper) SetTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	err := w.client.Set(key, value)
	if err != nil {
		return err
	}
	if ttl > 0 {
		w.client.SetTTL(key, ttl)
	}
	return nil
}
