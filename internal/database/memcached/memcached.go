package memcached

type Wrapper struct{}

func NewWrapper() (*Wrapper, error) {
	return &Wrapper{}, nil
}

func (w *Wrapper) Close() error {
	return nil
}
