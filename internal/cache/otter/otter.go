package otter

import (
	"context"
	"fmt"
	"time"

	"github.com/maypok86/otter/v2"
)

const defaultExpiration = time.Hour

type Wrapper struct {
	otter *otter.Cache[string, string]
}

func New() (*Wrapper, error) {
	opts := &otter.Options[string, string]{}
	opts.ExpiryCalculator = otter.ExpiryWriting[string, string](defaultExpiration)
	o, err := otter.New[string, string](opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise: %w", err)
	}
	return &Wrapper{o}, nil
}

func (w *Wrapper) Shutdown(_ context.Context) error {
	return nil
}

func (w *Wrapper) Get(_ context.Context, key string) (string, error) {
	v, present := w.otter.GetIfPresent(key)
	if !present {
		return "", nil
	}
	return v, nil
}

func (w *Wrapper) Set(_ context.Context, key string, value string) error {
	w.otter.Set(key, value)
	return nil
}

func (w *Wrapper) SetTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	w.otter.Set(key, value)
	w.otter.SetExpiresAfter(key, ttl)
	return nil
}

func (w *Wrapper) Invalidate(_ context.Context, key string) error {
	w.otter.Invalidate(key)
	return nil
}
