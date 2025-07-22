package domain

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

const (
	UserStatusActive = "active"
)

var (
	ErrEntityNotFound     = errors.New("entity not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user is not active")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID int      `json:"uid"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

type ContextAuthInfo struct {
	UserID int      `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}
