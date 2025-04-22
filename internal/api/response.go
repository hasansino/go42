package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func SendJSON(ctx *fiber.Ctx, data interface{}) error {
	return ctx.JSON(data)
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SendJSONError(ctx *fiber.Ctx, code int, msg string) error {
	return ctx.
		Status(code).
		JSON(Error{code, msg, nil})
}

func SendUnauthorisedError(ctx *fiber.Ctx) error {
	return SendJSONError(ctx, http.StatusUnauthorized,
		http.StatusText(http.StatusUnauthorized))
}

type ValidationError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Field   string `json:"field,omitempty"`
}

const (
	DefaultValidationErrMessage = "validation error"
	DefaultErrorDetailMessage   = "invalid value"
	DefaultErrorDetailCode      = "INVALID_VALUE"
)

func SendJSONValidationError(ctx *fiber.Ctx, vErrs ...ValidationError) error {
	message := DefaultValidationErrMessage
	return ctx.
		Status(http.StatusBadRequest).
		JSON(Error{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: vErrs,
		})
}
