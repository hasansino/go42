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

const DefaultValidationErrMessage = "validation error"

func SendJSONValidationError(ctx echo.Context, vErrs ...interface{}) error {
	return ctx.JSON(http.StatusBadRequest,
		Error{
			Code:    http.StatusBadRequest,
			Message: DefaultValidationErrMessage,
			Details: vErrs,
		})
}
