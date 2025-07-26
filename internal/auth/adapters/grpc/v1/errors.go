package adapter

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hasansino/go42/internal/auth/domain"
)

var (
	rpcErrorNotFound = status.New(codes.NotFound, "not found")
	rpcErrorInternal = status.New(codes.Internal, "internal error")
)

func (a *Adapter) processError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEntityNotFound):
		return rpcErrorNotFound.Err()
	default:
		return rpcErrorInternal.Err()
	}
}
