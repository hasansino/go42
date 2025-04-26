package sqlite

type Option func() (string, string)

func WithMode(mode string) Option {
	return func() (string, string) {
		return "mode", mode
	}
}
