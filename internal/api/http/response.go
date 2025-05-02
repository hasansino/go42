package http

import (
	"github.com/labstack/echo/v4"
)

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SendJSONError(ctx echo.Context, code int, msg string, errs ...interface{}) error {
	errorResponse := Error{Code: code, Message: msg}
	if len(errs) > 0 {
		errorResponse.Details = errs
	}
	return ctx.JSON(code, errorResponse)
}
