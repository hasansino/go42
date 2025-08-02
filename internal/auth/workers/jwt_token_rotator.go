package workers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"
)

const defaultSecretLength = 32 // 256 bits

type SecretRotationWorker struct {
	logger       *slog.Logger
	service      authService
	secretLength int
}

func NewSecretRotationWorker(
	service authService,
	opts ...SecretRotationWorkerOption,
) *SecretRotationWorker {
	r := &SecretRotationWorker{
		service:      service,
		secretLength: defaultSecretLength,
	}
	for _, o := range opts {
		o(r)
	}
	if r.logger == nil {
		r.logger = slog.New(slog.DiscardHandler)
	}
	return r
}

func (w *SecretRotationWorker) Run(ctx context.Context, interval time.Duration) {
	w.logger.InfoContext(ctx, "starting JWT secret rotation worker")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.run(ctx)
		}
	}
}

func (w *SecretRotationWorker) run(ctx context.Context) {
	newSecret, err := w.generateSecret()
	if err != nil {
		w.logger.ErrorContext(ctx, "failed to rotate JWT secret",
			slog.Any("error", err),
		)
		return
	}
	w.service.RotateJWTSecret(newSecret)
	w.logger.InfoContext(ctx, "JWT secret rotated successfully")
}

func (w *SecretRotationWorker) generateSecret() (string, error) {
	bytes := make([]byte, w.secretLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

type SecretRotationWorkerOption func(*SecretRotationWorker)

func SecretRotationWorkerWithLogger(logger *slog.Logger) SecretRotationWorkerOption {
	return func(o *SecretRotationWorker) {
		o.logger = logger
	}
}

func SecretRotationWorkerWithSecretLength(length int) SecretRotationWorkerOption {
	return func(o *SecretRotationWorker) {
		o.secretLength = length
	}
}
