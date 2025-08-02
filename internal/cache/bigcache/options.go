package bigcache

import (
	"time"

	"github.com/allegro/bigcache"
)

type Option func(*bigcache.Config)

func WithShards(shards int) Option {
	return func(cfg *bigcache.Config) {
		cfg.Shards = shards
	}
}
func WithLifeWindow(lifeWindow time.Duration) Option {
	return func(cfg *bigcache.Config) {
		cfg.LifeWindow = lifeWindow
	}
}

func WithMaxEntriesInWindow(maxEntriesInWindow int) Option {
	return func(cfg *bigcache.Config) {
		cfg.MaxEntriesInWindow = maxEntriesInWindow
	}
}

func WithMaxEntrySizeBytes(maxEntrySize int) Option {
	return func(cfg *bigcache.Config) {
		cfg.MaxEntrySize = maxEntrySize
	}
}

func WithHardMaxCacheSize(hardMaxCacheSize int) Option {
	return func(cfg *bigcache.Config) {
		cfg.HardMaxCacheSize = hardMaxCacheSize
	}
}

func WithVerbose(verbose bool) Option {
	return func(cfg *bigcache.Config) {
		cfg.Verbose = verbose
	}
}
