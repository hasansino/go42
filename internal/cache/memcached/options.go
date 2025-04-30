package memcached

import (
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type Option func(client *memcache.Client)

func WithTimeout(timeout time.Duration) Option {
	return func(client *memcache.Client) {
		client.Timeout = timeout
	}
}

func WithMaxIdleConns(maxIdleConns int) Option {
	return func(client *memcache.Client) {
		client.MaxIdleConns = maxIdleConns
	}
}
