package provider

import "time"

type Option func(p *Provider)

func WithCache(cache cache, ttl time.Duration) Option {
	return func(p *Provider) {
		p.cache = cache
		p.cacheTTL = ttl
	}
}
