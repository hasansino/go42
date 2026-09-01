package pgsql

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestWithQueryTimeoutURL(t *testing.T) {
	const dsn = "postgres://user:secret@localhost:5432/app?sslmode=disable"

	configuredDSN, err := withQueryTimeout(dsn, 1500*time.Microsecond)
	if err != nil {
		t.Fatalf("withQueryTimeout() error = %v", err)
	}
	parsedDSN, err := url.Parse(configuredDSN)
	if err != nil {
		t.Fatalf("parse configured DSN: %v", err)
	}
	if got := parsedDSN.Query().Get("statement_timeout"); got != "2" {
		t.Errorf("statement_timeout = %q, want 2", got)
	}
	if got := parsedDSN.Query().Get("sslmode"); got != "disable" {
		t.Errorf("sslmode = %q, want disable", got)
	}
}

func TestWithQueryTimeoutKeywordDSN(t *testing.T) {
	const dsn = "host=localhost dbname=app sslmode=disable"

	configuredDSN, err := withQueryTimeout(dsn, 2*time.Second)
	if err != nil {
		t.Fatalf("withQueryTimeout() error = %v", err)
	}
	if !strings.Contains(configuredDSN, "statement_timeout=2000") {
		t.Errorf("withQueryTimeout() = %q, want statement_timeout=2000", configuredDSN)
	}
}

func TestWithQueryTimeoutDisabled(t *testing.T) {
	const dsn = "postgres://user:secret@localhost:5432/app"

	configuredDSN, err := withQueryTimeout(dsn, 0)
	if err != nil {
		t.Fatalf("withQueryTimeout() error = %v", err)
	}
	if configuredDSN != dsn {
		t.Errorf("withQueryTimeout() = %q, want unchanged DSN", configuredDSN)
	}
}
