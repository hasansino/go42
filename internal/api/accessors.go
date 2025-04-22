package api

import "github.com/gofiber/fiber/v2"

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/accessors.go

// providerAccessor for all handler providers.
type providerAccessor interface {
	Register(r fiber.Router)
}
