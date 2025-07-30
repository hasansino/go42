package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/hasansino/go42/internal/auth"
	"github.com/hasansino/go42/internal/auth/domain"
	"github.com/hasansino/go42/internal/auth/models"
)

const (
	metadataKeyXApiToken = "x-api-key"
)

type authServiceAccessor interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	ValidateAPIToken(ctx context.Context, token string) (*models.Token, error)
}

func NewUnaryAuthInterceptor(authService authServiceAccessor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (interface{}, error) {
		authCtx, err := authenticateRequest(ctx, authService)
		if err != nil {
			return nil, err
		}
		return handler(authCtx, req)
	}
}

func NewStreamAuthInterceptor(authService authServiceAccessor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		authCtx, err := authenticateRequest(ss.Context(), authService)
		if err != nil {
			return err
		}
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          authCtx,
		}
		return handler(srv, wrappedStream)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

func authenticateRequest(ctx context.Context, authService authServiceAccessor) (context.Context, error) {
	token, err := extractFromGRPCMetadata(ctx, metadataKeyXApiToken)
	if err != nil {
		return nil, err
	}

	tokenInfo, err := authService.ValidateAPIToken(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid api token")
	}

	user, err := authService.GetUserByID(ctx, tokenInfo.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, "error retrieving user")
	}

	if !user.IsActive() {
		return nil, status.Error(codes.Unauthenticated, "user is not allowed to authenticate")
	}

	authInfo := domain.ContextAuthInfo{
		ID:   tokenInfo.ID,
		UUID: tokenInfo.UUID.String(),
		Type: domain.AuthenticationTypeApiToken,
	}
	authInfo.SetPermissions(tokenInfo.PermissionList())

	return auth.SetAuthToContext(ctx, authInfo), nil
}

func extractFromGRPCMetadata(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	apiKey := md.Get(key)
	if len(apiKey) > 0 {
		return apiKey[0], nil
	}
	return "", status.Error(codes.Unauthenticated, "missing metadata key: "+key)
}
