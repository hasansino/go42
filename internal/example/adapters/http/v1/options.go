package adapter

import "time"

type Option func(p *Adapter)

func WithCache(cache cache, ttl time.Duration) Option {
	return func(p *Adapter) {
		p.cache = cache
		p.cacheTTL = ttl
	}
}
