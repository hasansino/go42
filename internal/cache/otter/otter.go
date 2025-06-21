package otter

import (
	"context"
	"fmt"
	"time"

	"github.com/maypok86/otter/v2"
)

type Wrapper struct {
	otter *otter.Cache[string, string]
}

func New() (*Wrapper, error) {
	opts := &otter.Options[string, string]{}
	o, err := otter.New[string, string](opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise: %w", err)
	}
	return &Wrapper{o}, nil
}

func (w *Wrapper) Shutdown(_ context.Context) error {
	w.otter.CleanUp()
	return nil
}

func (w *Wrapper) Get(ctx context.Context, key string) (string, error) {
	loader := func(ctx context.Context, key string) (string, error) {
		return "", nil
	}
	return w.otter.Get(ctx, key, otter.LoaderFunc[string, string](loader))
}

func (w *Wrapper) Set(_ context.Context, key string, value string) error {
	w.otter.Set(key, value)
	return nil
}

func (w *Wrapper) SetTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	w.otter.SetExpiresAfter(key, ttl)
	return nil
}
