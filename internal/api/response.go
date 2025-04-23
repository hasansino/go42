package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func SendJSON(ctx echo.Context, data interface{}) error {
	return ctx.JSON(http.StatusOK, data)
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SendJSONError(ctx echo.Context, code int, msg string) error {
	return ctx.JSON(code, Error{code, msg, nil})
}

func SendUnauthorisedError(ctx echo.Context) error {
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

func SendJSONValidationError(ctx echo.Context, vErrs ...ValidationError) error {
	message := DefaultValidationErrMessage
	return ctx.JSON(http.StatusBadRequest,
		Error{
			Code:    http.StatusBadRequest,
			Message: message,
			Details: vErrs,
		})
}
