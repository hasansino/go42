package sqlite

type Option func() (string, string)

func WithMode(mode string) Option {
	return func() (string, string) {
		return "mode", mode
	}
}

func WithCacheMod(mode string) Option {
	return func() (string, string) {
		return "cache", mode
	}
}
