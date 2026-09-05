package adapter

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go42-dev/go42/internal/auth/domain"
)

type errorMappingTestCase struct {
	name        string
	err         error
	wantCode    codes.Code
	wantMessage string
}

func TestAdapterProcessError(t *testing.T) {
	testCases := []errorMappingTestCase{
		{
			name: "entity not found", err: domain.ErrEntityNotFound,
			wantCode: codes.NotFound, wantMessage: "not found",
		},
		{
			name: "wrapped entity not found", err: fmt.Errorf("lookup user: %w", domain.ErrEntityNotFound),
			wantCode: codes.NotFound, wantMessage: "not found",
		},
		{
			name: "user already exists", err: domain.ErrUserAlreadyExists,
			wantCode: codes.AlreadyExists, wantMessage: "user already exists",
		},
		{
			name: "invalid credentials", err: domain.ErrInvalidCredentials,
			wantCode: codes.InvalidArgument, wantMessage: "invalid credentials",
		},
		{
			name: "weak password", err: domain.ErrPasswordWeak,
			wantCode: codes.InvalidArgument, wantMessage: "password is too weak",
		},
		{
			name: "invalid pagination", err: domain.ErrInvalidPagination,
			wantCode: codes.InvalidArgument, wantMessage: "invalid pagination",
		},
		{
			name: "authentication unavailable", err: domain.ErrAuthenticationUnavailable,
			wantCode: codes.Unavailable, wantMessage: "authentication unavailable",
		},
		{
			name: "invalid token", err: domain.ErrInvalidToken,
			wantCode: codes.Unauthenticated, wantMessage: "invalid token",
		},
		{
			name: "unknown internal error", err: errors.New("database password leaked"),
			wantCode: codes.Internal, wantMessage: "internal error",
		},
	}
	adapter := &Adapter{}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := status.Convert(adapter.processError(testCase.err))
			if got.Code() != testCase.wantCode {
				t.Errorf("processError() code = %s, want %s", got.Code(), testCase.wantCode)
			}
			if got.Message() != testCase.wantMessage {
				t.Errorf("processError() message = %q, want %q", got.Message(), testCase.wantMessage)
			}
		})
	}
}
