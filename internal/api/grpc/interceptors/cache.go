package interceptors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	cacheHeaderName      = "x-cache"
	cacheHeaderValueHit  = "HIT"
	cacheHeaderValueMISS = "MISS"
)

type cacheAccessor interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Invalidate(ctx context.Context, key string) error
}

func NewUnaryServerCacheInterceptor(cache cacheAccessor, ttl time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if ttl == 0 {
			return handler(ctx, req)
		}

		key, err := cacheKeyFromUnaryRequest(ctx, info.FullMethod, req)
		if err != nil {
			slog.Default().
				With(slog.String("component", "grpc-interceptor-cache")).
				ErrorContext(ctx, "failed to generate cache key", slog.Any("error", err))
			return handler(ctx, req)
		}

		cached, err := cache.Get(ctx, key)
		if err != nil {
			slog.Default().
				With(slog.String("component", "grpc-interceptor-cache")).
				ErrorContext(ctx, "failed to fetch cached data", slog.Any("error", err))
			return handler(ctx, req)
		}

		if len(cached) > 0 {
			if err := grpc.SetHeader(ctx, metadata.Pairs(cacheHeaderName, cacheHeaderValueHit)); err != nil {
				slog.Default().
					With(slog.String("component", "grpc-interceptor-cache")).
					ErrorContext(ctx, "failed to set cache hit header", slog.Any("error", err))
			}

			resp := info.Server
			if resp == nil {
				return nil, fmt.Errorf("cannot determine response type")
			}

			if err := json.Unmarshal([]byte(cached), &resp); err != nil {
				slog.Default().
					With(slog.String("component", "grpc-interceptor-cache")).
					ErrorContext(ctx, "failed to unmarshal cached response", slog.Any("error", err))
				return handler(ctx, req)
			}

			return resp, nil
		}

		if err := grpc.SetHeader(ctx, metadata.Pairs(cacheHeaderName, cacheHeaderValueMISS)); err != nil {
			slog.Default().
				With(slog.String("component", "grpc-interceptor-cache")).
				ErrorContext(ctx, "failed to set cache miss header", slog.Any("error", err))
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		if resp != nil {
			data, err := json.Marshal(resp)
			if err != nil {
				slog.Default().
					With(slog.String("component", "grpc-interceptor-cache")).
					ErrorContext(ctx, "failed to marshal response for caching", slog.Any("error", err))
			} else {
				if err := cache.Set(ctx, key, string(data), ttl); err != nil {
					slog.Default().
						With(slog.String("component", "grpc-interceptor-cache")).
						ErrorContext(ctx, "failed to cache response", slog.Any("error", err))
				}
			}
		}

		return resp, err
	}
}

func NewUnaryClientCacheInterceptor(cache cacheAccessor, ttl time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if ttl == 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		key, err := cacheKeyFromUnaryRequest(ctx, method, req)
		if err != nil {
			slog.Default().
				With(slog.String("component", "grpc-client-interceptor-cache")).
				ErrorContext(ctx, "failed to generate cache key", slog.Any("error", err))
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		cached, err := cache.Get(ctx, key)
		if err != nil {
			slog.Default().
				With(slog.String("component", "grpc-client-interceptor-cache")).
				ErrorContext(ctx, "failed to fetch cached data", slog.Any("error", err))
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		if len(cached) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, cacheHeaderName, cacheHeaderValueHit)
			if err := json.Unmarshal([]byte(cached), reply); err != nil {
				slog.Default().
					With(slog.String("component", "grpc-client-interceptor-cache")).
					ErrorContext(ctx, "failed to unmarshal cached response", slog.Any("error", err))
				return invoker(ctx, method, req, reply, cc, opts...)
			}
			return nil
		}

		ctx = metadata.AppendToOutgoingContext(ctx, cacheHeaderName, cacheHeaderValueMISS)

		err = invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			return err
		}

		if reply != nil {
			data, err := json.Marshal(reply)
			if err != nil {
				slog.Default().
					With(slog.String("component", "grpc-client-interceptor-cache")).
					ErrorContext(ctx, "failed to marshal response for caching", slog.Any("error", err))
			} else {
				if err := cache.Set(ctx, key, string(data), ttl); err != nil {
					slog.Default().
						With(slog.String("component", "grpc-client-interceptor-cache")).
						ErrorContext(ctx, "failed to cache response", slog.Any("error", err))
				}
			}
		}

		return nil
	}
}

func cacheKeyFromUnaryRequest(ctx context.Context, method string, req any) (string, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s:%s", method, hex.EncodeToString(hash[:])), nil
}
