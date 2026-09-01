package interceptors

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	grpcmocks "github.com/hasansino/go42/internal/api/grpc/mocks"
)

const testClientMethod = "/payments.v1.PaymentService/Charge"

func TestExtractRateLimitKeyFromCtx(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		peer     *peer.Peer
		expected string
	}{
		{
			name:     "missing peer",
			expected: "",
		},
		{
			name:     "nil peer address",
			peer:     &peer.Peer{},
			expected: "",
		},
		{
			name: "IPv4 address",
			peer: &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("192.0.2.10"), Port: 54321},
			},
			expected: "192.0.2.10",
		},
		{
			name: "same IPv4 address with another port",
			peer: &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("192.0.2.10"), Port: 61482},
			},
			expected: "192.0.2.10",
		},
		{
			name: "IPv6 address",
			peer: &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("2001:db8::1"), Port: 54321},
			},
			expected: "2001:db8::1",
		},
		{
			name: "same IPv6 address with another port",
			peer: &peer.Peer{
				Addr: &net.TCPAddr{IP: net.ParseIP("2001:db8::1"), Port: 61482},
			},
			expected: "2001:db8::1",
		},
		{
			name: "non-TCP address",
			peer: &peer.Peer{
				Addr: &net.UnixAddr{Name: "/run/go42.sock", Net: "unix"},
			},
			expected: "/run/go42.sock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if tt.peer != nil {
				ctx = peer.NewContext(ctx, tt.peer)
			}

			assert.Equal(t, tt.expected, extractRateLimitKeyFromCtx(ctx))
		})
	}
}

func TestDefaultClientRateLimitKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	key := defaultClientRateLimitKey(ctx, "dns:///payments.example", testClientMethod)

	assert.Equal(
		t,
		"grpc-client|dns:///payments.example|/payments.v1.PaymentService/Charge",
		key,
	)
	assert.Equal(t, key, defaultClientRateLimitKey(ctx, "dns:///payments.example", testClientMethod))
	assert.NotEqual(t, key, defaultClientRateLimitKey(ctx, "dns:///email.example", testClientMethod))
	assert.NotEqual(
		t,
		key,
		defaultClientRateLimitKey(ctx, "dns:///payments.example", "/payments.v1.PaymentService/Refund"),
	)
}

func TestUnaryClientRateLimiterInterceptor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		allowed      bool
		limiterError error
		expectedCode codes.Code
		invoked      bool
	}{
		{
			name:    "allowed",
			allowed: true,
			invoked: true,
		},
		{
			name:         "denied",
			expectedCode: codes.ResourceExhausted,
		},
		{
			name:         "limiter unavailable",
			limiterError: errors.New("cache unavailable"),
			expectedCode: codes.Unavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			limiter := grpcmocks.NewMockrateLimiterAccessor(ctrl)
			limiter.EXPECT().
				Limit(gomock.Any(), "grpc-client|payments|"+testClientMethod).
				Return(tt.allowed, tt.limiterError)

			conn := newTestClientConn(t, "dns:///payments.example")
			interceptor := UnaryClientRateLimiterInterceptor(
				limiter,
				WithClientRateLimitScope("payments"),
			)

			invoked := false
			invoker := func(
				context.Context,
				string,
				interface{},
				interface{},
				*grpc.ClientConn,
				...grpc.CallOption,
			) error {
				invoked = true
				return nil
			}

			err := interceptor(
				context.Background(),
				testClientMethod,
				nil,
				nil,
				conn,
				invoker,
			)

			assert.Equal(t, tt.invoked, invoked)
			assert.Equal(t, tt.expectedCode, status.Code(err))
		})
	}
}

func TestStreamClientRateLimiterInterceptor(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	limiter := grpcmocks.NewMockrateLimiterAccessor(ctrl)
	limiter.EXPECT().
		Limit(gomock.Any(), "grpc-client|dns:///payments.example|"+testClientMethod).
		Return(false, nil)

	conn := newTestClientConn(t, "dns:///payments.example")
	interceptor := StreamClientRateLimiterInterceptor(limiter)

	invoked := false
	streamer := func(
		context.Context,
		*grpc.StreamDesc,
		*grpc.ClientConn,
		string,
		...grpc.CallOption,
	) (grpc.ClientStream, error) {
		invoked = true
		return nil, nil
	}

	stream, err := interceptor(
		context.Background(),
		&grpc.StreamDesc{},
		conn,
		testClientMethod,
		streamer,
	)

	assert.Nil(t, stream)
	assert.False(t, invoked)
	assert.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func newTestClientConn(t *testing.T, target string) *grpc.ClientConn {
	t.Helper()

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Close())
	})
	return conn
}
