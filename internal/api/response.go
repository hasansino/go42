package api

import (
	"github.com/labstack/echo/v4"
)

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SendJSONError(ctx echo.Context, code int, msg string, errs ...interface{}) error {
	return ctx.JSON(code, Error{code, msg, errs})
}
