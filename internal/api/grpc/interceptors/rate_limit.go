package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type rateLimiterAcessor interface {
	Limit(key any) bool
}

func UnaryServerRateLimiterInterceptor(limiter rateLimiterAcessor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if limiter == nil {
			return handler(ctx, req)
		}
		if !limiter.Limit(extractRateLimitKeyFromCtx(ctx)) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

func StreamServerRateLimiterInterceptor(limiter rateLimiterAcessor) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if limiter == nil {
			return handler(srv, stream)
		}
		if !limiter.Limit(extractRateLimitKeyFromCtx(stream.Context())) {
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(srv, stream)
	}
}

func UnaryClientRateLimiterInterceptor(limiter rateLimiterAcessor) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if limiter == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		if !limiter.Limit(extractRateLimitKeyFromCtx(ctx)) {
			return status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamClientRateLimiterInterceptor(limiter rateLimiterAcessor) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if limiter == nil {
			return streamer(ctx, desc, cc, method, opts...)
		}
		if !limiter.Limit(extractRateLimitKeyFromCtx(ctx)) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func extractRateLimitKeyFromCtx(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return ""
}
