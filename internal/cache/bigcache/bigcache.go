package bigcache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/allegro/bigcache"
)

type Wrapper struct {
	cache *bigcache.BigCache
}

func New(opts ...Option) (*Wrapper, error) {
	// defaults settings overwritten by options
	cfg := bigcache.Config{
		Shards:             1,
		LifeWindow:         5 * time.Minute,
		MaxEntriesInWindow: 1000,
		MaxEntrySize:       512 * 1024, // 512 KB
		HardMaxCacheSize:   100000,     // ~500 MB
		Verbose:            true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Verbose {
		cfg.Logger = slog.NewLogLogger(
			slog.Default().Handler().WithAttrs([]slog.Attr{
				slog.String("component", "bigcache"),
			}),
			slog.LevelDebug,
		)
	}

	cache, err := bigcache.NewBigCache(cfg)
	if err != nil {
		return nil, err
	}

	return &Wrapper{cache: cache}, nil
}

func (w *Wrapper) Shutdown(_ context.Context) error {
	return w.cache.Close()
}

func (w *Wrapper) Get(_ context.Context, key string) (string, error) {
	bytes, err := w.cache.Get(key)
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
			return "", nil
		}
		return "", err
	}
	return string(bytes), nil
}

func (w *Wrapper) Set(_ context.Context, key string, value string, _ time.Duration) error {
	return w.cache.Set(key, []byte(value))
}

func (w *Wrapper) Invalidate(_ context.Context, key string) error {
	return w.cache.Delete(key)
}
