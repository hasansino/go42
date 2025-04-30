package memcached

import (
	"context"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type Wrapper struct {
	client *memcache.Client
}

func New(hosts []string, opts ...Option) (*Wrapper, error) {
	client := memcache.New(hosts...)
	for _, opt := range opts {
		opt(client)
	}
	return &Wrapper{client: client}, client.Ping()
}

func (w *Wrapper) Close() error {
	return w.client.Close()
}

func (w *Wrapper) Get(_ context.Context, key string) (string, error) {
	item, err := w.client.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (w *Wrapper) Set(_ context.Context, key string, value string) error {
	return w.client.Set(&memcache.Item{Key: key, Value: []byte(value)})
}

func (w *Wrapper) SetTTL(_ context.Context, key string, value string, ttl time.Duration) error {
	return w.client.Set(&memcache.Item{Key: key, Value: []byte(value), Expiration: int32(ttl.Seconds())})
}
