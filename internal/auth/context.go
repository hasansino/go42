package auth

import (
	"context"

	"github.com/hasansino/go42/internal/auth/domain"
)

type contextKey string

const contextKeyAuthInfo contextKey = "auth_user_info"

// RetrieveAuthFromContext retrieve auth state from context.
// Returns pointer to copy of data, which can be mutated without side effects.
func RetrieveAuthFromContext(ctx context.Context) *domain.ContextAuthInfo {
	authInfo, ok := ctx.Value(contextKeyAuthInfo).(domain.ContextAuthInfo)
	if !ok {
		return nil
	}
	return &authInfo
}

// SetAuthToContext sets auth state to context.
func SetAuthToContext(ctx context.Context, authInfo domain.ContextAuthInfo) context.Context {
	return context.WithValue(ctx, contextKeyAuthInfo, authInfo)
}
