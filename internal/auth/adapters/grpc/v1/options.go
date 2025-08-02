package adapter

type Option func(a *Adapter)

func WithPermissionRegistry(registry permissionRegistry) Option {
	return func(p *Adapter) {
		p.permissionRegistry = registry
	}
}
