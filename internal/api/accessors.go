package api

import (
	"github.com/labstack/echo"
)

//go:generate mockgen -source $GOFILE -package mocks -destination mocks/accessors.go

// providerAccessor for all handler providers.
type providerAccessor interface {
	Register(r *echo.Group)
}
