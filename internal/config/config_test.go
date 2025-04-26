package config

import (
	"testing"

	"log/slog"
)

func TestLogger_Level(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{"DefaultInfoLevel", "info", slog.LevelInfo},
		{"DebugLevel", "debug", slog.LevelDebug},
		{"WarnLevel", "warn", slog.LevelWarn},
		{"ErrorLevel", "error", slog.LevelError},
		{"InfoWithPositiveOffset", "info+2", slog.LevelInfo + 2},
		{"InfoWithNegativeOffset", "info-2", slog.LevelInfo - 2},
		{"DebugWithPositiveOffset", "debug+1", slog.LevelDebug + 1},
		{"WarnWithNegativeOffset", "warn-1", slog.LevelWarn - 1},
		{"InvalidLevelDefaultsToInfo", "invalid", slog.LevelInfo},
		{"EmptyLevelDefaultsToInfo", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Logger{LogLevel: tt.logLevel}
			if got := logger.Level(); got != tt.expected {
				t.Errorf("Logger.Level() = %v, want %v", got, tt.expected)
			}
		})
	}
}
