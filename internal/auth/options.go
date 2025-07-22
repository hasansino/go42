package auth

import (
	"log/slog"
	"time"
)

type Option func(*Service)

func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

func WithJWTSecret(secret string) Option {
	return func(s *Service) {
		s.jwtSecret = secret
	}
}

func WithAccessTokenTTL(ttl time.Duration) Option {
	return func(s *Service) {
		s.accessTokenTTL = ttl
	}
}

func WithRefreshTokenTTL(ttl time.Duration) Option {
	return func(s *Service) {
		s.refreshTokenTTL = ttl
	}
}
