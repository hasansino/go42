package auth

import (
	"context"

	"github.com/hasansino/go42/internal/auth/domain"
)

type contextKey string

const contextKeyAuthInfo contextKey = "auth_user_info"

func RetrieveAuthFromContext(ctx context.Context) *domain.ContextAuthInfo {
	authInfo, ok := ctx.Value(contextKeyAuthInfo).(*domain.ContextAuthInfo)
	if !ok {
		return nil
	}
	return authInfo
}

func SetAuthToContext(ctx context.Context, authInfo *domain.ContextAuthInfo) context.Context {
	return context.WithValue(ctx, contextKeyAuthInfo, authInfo)
}
