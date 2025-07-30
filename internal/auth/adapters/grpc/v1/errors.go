package adapter

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hasansino/go42/internal/auth/domain"
)

func (a *Adapter) processError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEntityNotFound):
		return status.Error(codes.NotFound, "not found")
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.InvalidArgument, "invalid credentials")
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "invalid token")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
