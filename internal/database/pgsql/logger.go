package pgsql

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
)

type SlogAdapter struct {
	logger *slog.Logger
}

func NewSlogAdapter(l *slog.Logger) *SlogAdapter {
	return &SlogAdapter{logger: l}
}

func (s *SlogAdapter) LogMode(_ logger.LogLevel) logger.Interface {
	return s
}

func (s *SlogAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	s.logger.InfoContext(ctx, msg, slog.Any("data", data))
}

func (s *SlogAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	s.logger.WarnContext(ctx, msg, slog.Any("data", data))
}

func (s *SlogAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	s.logger.ErrorContext(ctx, msg, slog.Any("data", data))
}

func (s *SlogAdapter) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (string, int64),
	err error,
) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []interface{}{
		"elapsed", elapsed,
		"rows", rows,
	}
	if err != nil {
		s.logger.ErrorContext(ctx, "SQL Trace", append(fields, "error", err, "sql", sql)...)
	} else {
		s.logger.InfoContext(ctx, "SQL Trace", append(fields, "sql", sql)...)
	}
}
