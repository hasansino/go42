package tools

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	HeaderAuthorization     = "Authorization"
	HeaderValueBearerPrefix = "Bearer "
)

// ExtractBearerTokenFromRequest extracts the Bearer token from the Authorization header.
func ExtractBearerTokenFromRequest(req *http.Request) (string, error) {
	if req == nil {
		return "", fmt.Errorf("request is nil")
	}
	authHeader := req.Header.Get(HeaderAuthorization)
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	if !strings.HasPrefix(authHeader, HeaderValueBearerPrefix) {
		return "", fmt.Errorf("invalid Authorization header format, expected Bearer token")
	}
	token := authHeader[len(HeaderValueBearerPrefix):]
	if token == "" {
		return "", fmt.Errorf("empty token in Authorization header")
	}
	return token, nil
}
