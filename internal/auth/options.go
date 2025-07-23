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

func WithJWTAccessTokenTTL(ttl time.Duration) Option {
	return func(s *Service) {
		s.accessTokenTTL = ttl
	}
}

func WithJWTRefreshTokenTTL(ttl time.Duration) Option {
	return func(s *Service) {
		s.refreshTokenTTL = ttl
	}
}

func WithJWTIssuer(issuer string) Option {
	return func(s *Service) {
		s.jwtIssuer = issuer
	}
}

func WithJWTAudience(audience []string) Option {
	return func(s *Service) {
		s.jwtAudience = audience
	}
}
