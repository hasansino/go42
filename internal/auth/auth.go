package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
	outboxDomain "github.com/hasansino/go42/internal/outbox/domain"
)

const (
	cacheKeyInvalidatedToken   = "auth_invalidated_"
	cacheValueInvalidatedToken = "_"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type repository interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	AssignRoleToUser(ctx context.Context, userID int, role string) error
	GetToken(ctx context.Context, token string) (*models.Token, error)
}

type cache interface {
	Get(ctx context.Context, key string) (string, error)
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
}

type outboxService interface {
	NewOutboxMessage(ctx context.Context, topic string, msg *outboxDomain.Message) error
}

type Service struct {
	logger        *slog.Logger
	repository    repository
	cache         cache
	outboxService outboxService

	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	jwtIssuer       string
	jwtAudience     []string

	lastUsedTokens   map[int]time.Time
	lastUsedTokensMu sync.Mutex
}

func NewService(
	repository repository,
	outboxService outboxService,
	cache cache,
	opts ...Option,
) *Service {
	s := &Service{
		repository:     repository,
		outboxService:  outboxService,
		cache:          cache,
		lastUsedTokens: make(map[int]time.Time),
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.logger == nil {
		s.logger = slog.New(slog.DiscardHandler)
	}
	return s
}

func (s *Service) SignUp(ctx context.Context, email string, password string) (*models.User, error) {
	user := &models.User{
		UUID:   uuid.New(),
		Email:  email,
		Status: domain.UserStatusActive,
	}
	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	err := s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repository.CreateUser(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		if err := s.repository.AssignRoleToUser(txCtx, user.ID, domain.RBACRoleUser); err != nil {
			return fmt.Errorf("failed to assign user role: %w", err)
		}
		event := outboxDomain.Message{
			AggregateID:   user.ID,
			AggregateType: domain.EventTypeAuthSignUp,
		}
		if err := s.sendEvent(ctx, domain.TopicNameAuthEvents, event); err != nil {
			s.logger.ErrorContext(
				ctx, "failed to send event: %w",
				slog.String("topic", domain.TopicNameAuthEvents),
				slog.Any("event", event),
				slog.Any("error", err),
			)
			// assuming events are non-critical, do not fail transaction
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (*domain.Tokens, error) {
	user, err := s.repository.GetUserByEmail(
		ctx,
		strings.ToLower(strings.TrimSpace(email)),
	)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive() {
		return nil, domain.ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password.V), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(user.UUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	event := outboxDomain.Message{
		AggregateID:   user.ID,
		AggregateType: domain.EventTypeAuthLogin,
	}
	if err := s.sendEvent(ctx, domain.TopicNameAuthEvents, event); err != nil {
		s.logger.ErrorContext(
			ctx, "failed to send event: %w",
			slog.String("topic", domain.TopicNameAuthEvents),
			slog.Any("event", event),
			slog.Any("error", err),
		)
	}

	return tokens, nil
}

func (s *Service) Refresh(ctx context.Context, token string) (*domain.Tokens, error) {
	claims, err := s.ValidateJWTToken(ctx, token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.repository.GetUserByUUID(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		if err := s.InvalidateJWTToken(ctx, token, claims.ExpiresAt.Time); err != nil {
			s.logger.ErrorContext(
				ctx, "failed to invalidate token",
				slog.Any("err", err),
			)
		}
		return nil, domain.ErrInvalidToken
	}

	return s.generateTokens(user.UUID.String())
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	accessTokenClaims, err := s.ValidateJWTToken(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("invalid access token: %w", err)
	}
	refreshTokenClaims, err := s.ValidateJWTToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}

	if err := s.InvalidateJWTToken(ctx, accessToken, accessTokenClaims.ExpiresAt.Time); err != nil {
		return fmt.Errorf("failed to invalidate access token: %w", err)
	}

	if err := s.InvalidateJWTToken(ctx, refreshToken, refreshTokenClaims.ExpiresAt.Time); err != nil {
		return fmt.Errorf("failed to invalidate refresh token: %w", err)
	}

	user, err := s.repository.GetUserByUUID(ctx, accessTokenClaims.Subject)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	event := outboxDomain.Message{
		AggregateID:   user.ID,
		AggregateType: domain.EventTypeAuthLogout,
	}
	if err := s.sendEvent(ctx, domain.TopicNameAuthEvents, event); err != nil {
		s.logger.ErrorContext(
			ctx, "failed to send event: %w",
			slog.String("topic", domain.TopicNameAuthEvents),
			slog.Any("event", event),
			slog.Any("error", err),
		)
	}

	return nil
}

// ----

func (s *Service) UpdateUser(ctx context.Context, id int, fn func(*models.User) error) error {
	return s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.repository.GetUserByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if err := fn(user); err != nil {
			if errors.Is(err, domain.ErrNothingToUpdate) {
				return nil
			}
			return fmt.Errorf("updateFn failed: %w", err)
		}
		if err := s.repository.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		event := outboxDomain.Message{
			AggregateID:   user.ID,
			AggregateType: domain.EventTypeUserUpdate,
		}
		if err := s.sendEvent(ctx, domain.TopicNameAuthEvents, event); err != nil {
			s.logger.ErrorContext(
				ctx, "failed to send event: %w",
				slog.String("topic", domain.TopicNameAuthEvents),
				slog.Any("event", event),
				slog.Any("error", err),
			)
			// assuming events are non-critical, do not fail transaction
		}
		return nil
	})
}

func (s *Service) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return s.repository.GetUserByID(ctx, id)
}

func (s *Service) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	return s.repository.GetUserByUUID(ctx, uuid)
}

// ----

func (s *Service) ValidateJWTToken(ctx context.Context, token string) (*jwt.RegisteredClaims, error) {
	t, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*jwt.RegisteredClaims)
	if !ok || !t.Valid {
		return nil, domain.ErrInvalidToken
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, domain.ErrInvalidToken
	}

	tokenHash := sha256.New().Sum([]byte(token))
	v, err := s.cache.Get(ctx, cacheKeyInvalidatedToken+string(tokenHash))
	if err != nil {
		return nil, fmt.Errorf("failed to access cache: %w", err)
	}

	if v == cacheValueInvalidatedToken {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) InvalidateJWTToken(ctx context.Context, token string, until time.Time) error {
	return s.cache.SetTTL(
		ctx,
		cacheKeyInvalidatedToken+string(sha256.New().Sum([]byte(token))),
		cacheValueInvalidatedToken,
		time.Until(until)+1,
	)
}

func (s *Service) generateTokens(userUUID string) (*domain.Tokens, error) {
	accessToken, err := s.generateAccessToken(userUUID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(userUUID)
	if err != nil {
		return nil, err
	}

	return &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) generateAccessToken(userUUID string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Audience:  s.jwtAudience,
		Issuer:    s.jwtIssuer,
		Subject:   userUUID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
	}).SignedString([]byte(s.jwtSecret))
}

func (s *Service) generateRefreshToken(userUUID string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Audience:  s.jwtAudience,
		Issuer:    s.jwtIssuer,
		Subject:   userUUID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenTTL)),
	}).SignedString([]byte(s.jwtSecret))
}

// ----

func (s *Service) ValidateAPIToken(ctx context.Context, token string) (*models.Token, error) {
	apiToken, err := s.repository.GetToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid api token: %w", err)
	}

	if apiToken.ExpiresAt.Valid && apiToken.ExpiresAt.V.Before(time.Now()) {
		return nil, fmt.Errorf("expired api token: %w", err)
	}

	// a little sacrifice for the sake of ux
	go func() {
		s.lastUsedTokensMu.Lock()
		defer s.lastUsedTokensMu.Unlock()
		s.lastUsedTokens[apiToken.ID] = time.Now()
	}()

	return apiToken, nil
}

func (s *Service) RecentlyUsedTokens() map[int]time.Time {
	s.lastUsedTokensMu.Lock()
	defer s.lastUsedTokensMu.Unlock()
	tokens := s.lastUsedTokens
	s.lastUsedTokens = make(map[int]time.Time)
	return tokens
}

// ----

func (s *Service) sendEvent(ctx context.Context, topic string, outboxMessage outboxDomain.Message) error {
	err := s.outboxService.NewOutboxMessage(ctx, topic, &outboxMessage)
	if err != nil {
		return fmt.Errorf("failed to send outbox message: %w", err)
	}
	return nil
}
