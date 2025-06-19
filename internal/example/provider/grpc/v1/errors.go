package provider

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hasansino/go42/internal/example/domain"
)

var (
	rpcErrorNotFound = status.New(codes.NotFound, "not found")
	rpcErrorConflict = status.New(codes.AlreadyExists, "already exists")
	rpcErrorInternal = status.New(codes.Internal, "internal error")
)

func (p *Provider) processError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return rpcErrorNotFound.Err()
	case errors.Is(err, domain.ErrAlreadyExists):
		return rpcErrorConflict.Err()
	default:
		return rpcErrorInternal.Err()
	}
}
