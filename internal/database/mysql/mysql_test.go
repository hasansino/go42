package mysql

import (
	"strings"
	"testing"
	"time"

	libMysql "github.com/go-sql-driver/mysql"
)

func TestWithQueryTimeout(t *testing.T) {
	const dsn = "user:secret@tcp(localhost:3306)/app?parseTime=true&charset=utf8mb4"

	configuredDSN, err := withQueryTimeout(dsn, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("withQueryTimeout() error = %v", err)
	}
	config, err := libMysql.ParseDSN(configuredDSN)
	if err != nil {
		t.Fatalf("parse configured DSN: %v", err)
	}
	if config.ReadTimeout != 1500*time.Millisecond {
		t.Errorf("read timeout = %s, want 1.5s", config.ReadTimeout)
	}
	if config.WriteTimeout != 1500*time.Millisecond {
		t.Errorf("write timeout = %s, want 1.5s", config.WriteTimeout)
	}
	if !config.ParseTime || !strings.Contains(configuredDSN, "charset=utf8mb4") {
		t.Errorf("existing DSN settings were not preserved: %#v", config)
	}
}

func TestWithQueryTimeoutDisabled(t *testing.T) {
	const dsn = "user:secret@tcp(localhost:3306)/app?parseTime=true"

	configuredDSN, err := withQueryTimeout(dsn, 0)
	if err != nil {
		t.Fatalf("withQueryTimeout() error = %v", err)
	}
	if configuredDSN != dsn {
		t.Errorf("withQueryTimeout() = %q, want unchanged DSN", configuredDSN)
	}
}
