package middleware

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
)

var DefaultSkipper = func(c echo.Context) bool {
	if c.Path() == "/health" || c.Path() == "/metrics" {
		return true
	}
	return false
}

// ---

type responseRecorder struct {
	http.ResponseWriter
	status int

	write bool
	body  bytes.Buffer
	size  int
}

func newResponseRecorder(w http.ResponseWriter, write bool) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, write: write}
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.write {
		r.body.Write(b)
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

// GetUnderlyingWriter returns the underlying http.ResponseWriter
// This is needed for WebSocket upgrades that require http.Hijacker interface
func (r *responseRecorder) GetUnderlyingWriter() http.ResponseWriter {
	return r.ResponseWriter
}
