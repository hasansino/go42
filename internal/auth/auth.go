package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/mocks.go

type repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByUUID(ctx context.Context, uuid string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	AssignRoleToUser(ctx context.Context, userID int, role string) error
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type Service struct {
	logger          *slog.Logger
	repository      repository
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	jwtIssuer       string
	jwtAudience     []string
}

func NewService(repository repository, opts ...Option) *Service {
	s := &Service{
		repository: repository,
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
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		UUID:     uuid.New(),
		Email:    email,
		Password: sql.Null[string]{Valid: true, V: string(hash)},
		Status:   domain.UserStatusActive,
	}

	err = s.repository.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repository.CreateUser(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		if err := s.repository.AssignRoleToUser(txCtx, user.ID, domain.RBACRoleUser); err != nil {
			return fmt.Errorf("failed to assign user role: %w", err)
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

	if user.Status != domain.UserStatusActive {
		return nil, domain.ErrUserInactive
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password.V), []byte(password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}

	tokens, err := s.generateTokens(user.UUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return tokens, nil
}

func (s *Service) Refresh(ctx context.Context, token string) (*domain.Tokens, error) {
	claims, err := s.ValidateToken(token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.repository.GetUserByUUID(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}

	if user.Status != domain.UserStatusActive {
		return nil, domain.ErrUserInactive
	}

	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}

	return s.generateTokens(user.UUID.String())
}

func (s *Service) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	return s.repository.GetUserByID(ctx, id)
}

func (s *Service) GetUserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	return s.repository.GetUserByUUID(ctx, uuid)
}

func (s *Service) ValidateToken(token string) (*jwt.RegisteredClaims, error) {
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
		return nil, domain.ErrTokenExpired
	}

	return claims, nil
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
