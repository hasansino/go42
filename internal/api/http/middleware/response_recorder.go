package middleware

import (
	"bytes"
	"net/http"
)

type responseRecorder struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
	size   int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}
