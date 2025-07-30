package adapter

import "time"

type Option func(a *Adapter)

func WithCache(cache cache, ttl time.Duration) Option {
	return func(p *Adapter) {
		p.cache = cache
		p.cacheTTL = ttl
	}
}

func WithPermissionRegistry(registry permissionRegistry) Option {
	return func(p *Adapter) {
		p.permissionRegistry = registry
	}
}
