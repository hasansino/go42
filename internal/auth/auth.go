package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"

	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
	chatDomain "github.com/hasansino/go42/internal/chat/domain"
	outboxDomain "github.com/hasansino/go42/internal/outbox/domain"
	"github.com/hasansino/go42/internal/tools"
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
	DeleteUser(ctx context.Context, user *models.User) error
	ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	AssignRoleToUser(ctx context.Context, userID int, role string) error

	GetToken(ctx context.Context, hashedToken string) (*models.Token, error)
}

type cache interface {
	Get(ctx context.Context, key string) (string, error)
	SetTTL(ctx context.Context, key string, value string, ttl time.Duration) error
}

type outboxService interface {
	NewOutboxMessage(ctx context.Context, topic string, msg *outboxDomain.Message) error
}

// jwtSecret holds the SHA256 hash of the secret and the secret itself.
// The SHA256 hash is used as the KID (Key ID) in JWT tokens to
// allow for key rotation.
type jwtSecret struct {
	sha256 string
	secret string
}

type Service struct {
	logger        *slog.Logger
	repository    repository
	cache         cache
	outboxService outboxService

	jwtSecrets   []jwtSecret
	jwtSecretsMu sync.RWMutex

	jwtIssuer   string
	jwtAudience []string

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	minPasswordEntropyBits int

	tokensUsedChan chan domain.TokenWasUsed
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
		jwtSecrets:     make([]jwtSecret, 0, 2),
		tokensUsedChan: make(chan domain.TokenWasUsed, tools.BufferSize4096),
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
	if err := s.CheckPasswordStrength(password); err != nil {
		return nil, domain.ErrPasswordWeak
	}

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
		if err := s.sendEvent(txCtx, domain.TopicNameAuthEvents, event); err != nil {
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

func (s *Service) CreateUser(ctx context.Context, data *domain.CreateUserData) (*models.User, error) {
	if err := s.CheckPasswordStrength(data.Password); err != nil {
		return nil, domain.ErrPasswordWeak
	}

	user := &models.User{
		UUID:   uuid.New(),
		Email:  data.Email,
		Status: domain.UserStatusActive,
	}

	if err := user.SetPassword(data.Password); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	err := s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		err := s.repository.CreateUser(txCtx, user)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		if err := s.repository.AssignRoleToUser(txCtx, user.ID, domain.RBACRoleUser); err != nil {
			return fmt.Errorf("failed to assign user role: %w", err)
		}
		event := outboxDomain.Message{
			AggregateID:   user.ID,
			AggregateType: domain.EventTypeUserCreate,
		}
		err = s.sendEvent(txCtx, domain.TopicNameAuthEvents, event)
		if err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, uuid string, data *domain.UpdateUserData) error {
	return s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.repository.GetUserByUUID(txCtx, uuid)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		var doUpdate bool

		if data.Email != nil {
			if *data.Email != user.Email {
				doUpdate = true
				user.Email = *data.Email
			}
		}
		if data.Password != nil {
			doUpdate = true
			if err := s.CheckPasswordStrength(*data.Password); err != nil {
				return domain.ErrPasswordWeak
			}
			if err := user.SetPassword(*data.Password); err != nil {
				return fmt.Errorf("failed to set password: %w", err)
			}
		}

		if !doUpdate {
			return nil
		}

		if err := s.repository.UpdateUser(txCtx, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		event := outboxDomain.Message{
			AggregateID:   user.ID,
			AggregateType: domain.EventTypeUserUpdate,
		}
		if err := s.sendEvent(txCtx, domain.TopicNameAuthEvents, event); err != nil {
			s.logger.ErrorContext(
				txCtx, "failed to send event: %w",
				slog.String("topic", domain.TopicNameAuthEvents),
				slog.Any("event", event),
				slog.Any("error", err),
			)
			// assuming events are non-critical, do not fail transaction
		}

		return nil
	})
}

func (s *Service) DeleteUser(ctx context.Context, uuid string) error {
	return s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err := s.repository.GetUserByUUID(txCtx, uuid)
		if err != nil {
			return fmt.Errorf("failed to get user by id: %w", err)
		}
		err = s.repository.DeleteUser(txCtx, user)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
		event := outboxDomain.Message{
			AggregateID:   user.ID,
			AggregateType: domain.EventTypeUserDelete,
		}
		err = s.sendEvent(txCtx, domain.TopicNameAuthEvents, event)
		if err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
		return nil
	})
}

func (s *Service) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return s.repository.ListUsers(ctx, limit, offset)
}

func (s *Service) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return s.repository.GetUserByID(ctx, id)
}

func (s *Service) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	return s.repository.GetUserByUUID(ctx, uuid)
}

// ----

func (s *Service) RotateJWTSecret(newSecret string) {
	s.jwtSecretsMu.Lock()
	defer s.jwtSecretsMu.Unlock()

	hash := sha256.Sum256([]byte(newSecret))
	newJWTSecret := jwtSecret{
		sha256: hex.EncodeToString(hash[:]),
		secret: newSecret,
	}

	s.jwtSecrets = append(s.jwtSecrets, newJWTSecret)

	// only 3 secrets are kept in memory
	if len(s.jwtSecrets) > 3 {
		s.jwtSecrets = s.jwtSecrets[len(s.jwtSecrets)-3:]
	}
}

func (s *Service) ValidateJWTToken(ctx context.Context, token string) (*jwt.RegisteredClaims, error) {
	t, err := jwt.ParseWithClaims(token, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		s.jwtSecretsMu.RLock()
		defer s.jwtSecretsMu.RUnlock()
		for _, secret := range s.jwtSecrets {
			if kid == secret.sha256 {
				return []byte(secret.secret), nil
			}
		}
		return []byte(s.jwtSecrets[len(s.jwtSecrets)-1].secret), nil
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

	if v, err := s.cache.Get(ctx, cacheKeyInvalidatedToken+strToSHA256(token)); err != nil {
		s.logger.ErrorContext(
			ctx, "failed to fetch cache: %w",
			slog.Any("error", err))
	} else if v == cacheValueInvalidatedToken {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) InvalidateJWTToken(ctx context.Context, token string, until time.Time) error {
	return s.cache.SetTTL(
		ctx,
		cacheKeyInvalidatedToken+strToSHA256(token),
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
	s.jwtSecretsMu.RLock()
	token := s.jwtSecrets[len(s.jwtSecrets)-1]
	s.jwtSecretsMu.RUnlock()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, domain.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Audience:  s.jwtAudience,
			Issuer:    s.jwtIssuer,
			Subject:   userUUID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
		},
		KID: token.sha256,
	}).SignedString([]byte(token.secret))
}

func (s *Service) generateRefreshToken(userUUID string) (string, error) {
	s.jwtSecretsMu.RLock()
	token := s.jwtSecrets[len(s.jwtSecrets)-1]
	s.jwtSecretsMu.RUnlock()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, domain.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Audience:  s.jwtAudience,
			Issuer:    s.jwtIssuer,
			Subject:   userUUID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenTTL)),
		},
		KID: token.sha256,
	}).SignedString([]byte(token.secret))
}

// ----

func (s *Service) ValidateAPIToken(ctx context.Context, token string) (*models.Token, error) {
	apiToken, err := s.repository.GetToken(ctx, strToSHA256(token))
	if err != nil {
		return nil, fmt.Errorf("invalid api token: %w", err)
	}

	if apiToken.ExpiresAt.Valid && apiToken.ExpiresAt.V.Before(time.Now()) {
		return nil, fmt.Errorf("expired api token: %w", err)
	}

	select {
	case s.tokensUsedChan <- domain.TokenWasUsed{ID: apiToken.ID, When: time.Now()}:
	default:
		// if channel is full, we discard payload and record warning
		s.logger.WarnContext(ctx, "auth.tokensUsedChan overflow")
	}

	return apiToken, nil
}

func (s *Service) RecentlyUsedTokensChan() <-chan domain.TokenWasUsed {
	return s.tokensUsedChan
}

// ----

func (s *Service) sendEvent(ctx context.Context, topic string, outboxMessage outboxDomain.Message) error {
	err := s.outboxService.NewOutboxMessage(ctx, topic, &outboxMessage)
	if err != nil {
		return fmt.Errorf("failed to send outbox message: %w", err)
	}
	return nil
}

func (s *Service) CheckPasswordStrength(password string) error {
	return passwordvalidator.Validate(password, float64(s.minPasswordEntropyBits))
}

func strToSHA256(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ValidateTokenForChat validates a JWT token and returns basic user information for chat systems
// This implements the AuthService interface defined in chat domain
func (s *Service) ValidateTokenForChat(ctx context.Context, token string) (chatDomain.UserInfo, error) {
	// Validate the JWT token using the auth service
	claims, err := s.ValidateJWTToken(ctx, token)
	if err != nil {
		return chatDomain.UserInfo{}, err
	}

	// Get user details using the subject (user UUID) from claims
	user, err := s.GetUserByUUID(ctx, claims.Subject)
	if err != nil {
		return chatDomain.UserInfo{}, err
	}

	// Convert UUID to string for chat domain
	userUUIDStr := user.UUID.String()

	// Convert to chat domain UserInfo (hiding sensitive data like email, ID)
	return chatDomain.UserInfo{
		UUID:     userUUIDStr,
		Username: user.Email, // Using email as username for now
		JoinedAt: time.Now(),
	}, nil
}
