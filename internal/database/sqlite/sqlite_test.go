package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddConnectionOptions(t *testing.T) {
	tests := []struct {
		name     string
		dbPath   string
		connOpts []ConnectionOption
		expected string
	}{
		{
			name:     "No connection options",
			dbPath:   "test.db",
			connOpts: []ConnectionOption{},
			expected: "test.db",
		},
		{
			name:   "Single connection option",
			dbPath: "test.db",
			connOpts: []ConnectionOption{
				{Key: "mode", Value: "memory"},
			},
			expected: "test.db?mode=memory",
		},
		{
			name:   "Multiple connection options",
			dbPath: "test.db",
			connOpts: []ConnectionOption{
				{Key: "mode", Value: "memory"},
				{Key: "cache", Value: "shared"},
			},
			expected: "test.db?mode=memory&cache=shared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddConnectionOptions(tt.dbPath, tt.connOpts)
			assert.Equal(t, tt.expected, result)
		})
	}
}
