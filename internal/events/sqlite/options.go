package sqlite

import (
	"log/slog"
	"time"

	wmsqlite "github.com/ThreeDotsLabs/watermill-sqlite/wmsqlitemodernc"
)

type Option func(*SQLite, *wmsqlite.PublisherOptions, *wmsqlite.SubscriberOptions)

func WithLogger(logger *slog.Logger) Option {
	return func(s *SQLite, pubCfg *wmsqlite.PublisherOptions, subCfg *wmsqlite.SubscriberOptions) {
		s.logger = logger
	}
}

func WithConsumerGroup(group string) Option {
	return func(s *SQLite, pubCfg *wmsqlite.PublisherOptions, subCfg *wmsqlite.SubscriberOptions) {
		subCfg.ConsumerGroupMatcher = wmsqlite.NewStaticConsumerGroupMatcher(group)
	}
}

func WithBatchSize(size int) Option {
	return func(s *SQLite, pubCfg *wmsqlite.PublisherOptions, subCfg *wmsqlite.SubscriberOptions) {
		subCfg.BatchSize = size
	}
}

func WithPollInterval(interval time.Duration) Option {
	return func(s *SQLite, pubCfg *wmsqlite.PublisherOptions, subCfg *wmsqlite.SubscriberOptions) {
		subCfg.PollInterval = interval
	}
}

func WithLockTimeout(timeout time.Duration) Option {
	return func(s *SQLite, pubCfg *wmsqlite.PublisherOptions, subCfg *wmsqlite.SubscriberOptions) {
		subCfg.LockTimeout = timeout
	}
}

func WithAckDeadline(deadline time.Duration) Option {
	return func(s *SQLite, pubCfg *wmsqlite.PublisherOptions, subCfg *wmsqlite.SubscriberOptions) {
		subCfg.AckDeadline = &deadline
	}
}
