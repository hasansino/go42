package local

type config struct {
	capacity     uint64
	maxCostBytes uint64
}

type Option func(*config)

// WithCapacity sets the maximum number of entries. Zero disables the limit.
func WithCapacity(capacity uint64) Option {
	return func(cfg *config) {
		cfg.capacity = capacity
	}
}

// WithMaxCostBytes sets the maximum combined size of cached keys and values.
// Zero disables the limit; cache bookkeeping overhead is not included.
func WithMaxCostBytes(maxCostBytes uint64) Option {
	return func(cfg *config) {
		cfg.maxCostBytes = maxCostBytes
	}
}
