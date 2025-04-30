package redis

import "context"

type Wrapper struct{}

func New() (*Wrapper, error) {
	return &Wrapper{}, nil
}

func (w *Wrapper) Close() error {
	return nil
}

func (w *Wrapper) Get(ctx context.Context, key string, value interface{}) error {
	return nil
}

func (w *Wrapper) Set(ctx context.Context, key string, value interface{}) error {
	return nil
}
