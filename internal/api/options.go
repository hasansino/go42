package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Option func(f *fiber.Config)

func WithReadTimeout(d time.Duration) Option {
	return func(f *fiber.Config) {
		f.ReadTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(f *fiber.Config) {
		f.WriteTimeout = d
	}
}
