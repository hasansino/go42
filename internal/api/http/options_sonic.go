package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/labstack/echo/v4"
)

// sonicJSONSerializer is alternative json serializer
// @see https://github.com/bytedance/sonic
type sonicJSONSerializer struct {
	api sonic.API
}

func (s sonicJSONSerializer) Serialize(c echo.Context, i any, indent string) error {
	enc := s.api.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

func (s sonicJSONSerializer) Deserialize(c echo.Context, i any) error {
	err := s.api.NewDecoder(c.Request().Body).Decode(i)
	if err == nil {
		return nil
	}

	if ute := (*json.UnmarshalTypeError)(nil); errors.As(err, &ute) {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf(
				"Unmarshal error: got=%s, expected=%v, field=%s, offset=%d",
				ute.Value, ute.Type, ute.Field, ute.Offset,
			),
		).SetInternal(err)
	}

	if se := (*json.SyntaxError)(nil); errors.As(err, &se) {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf(
				"Syntax error: offset=%d, error=%s",
				se.Offset, se.Error(),
			),
		).SetInternal(err)
	}

	return err
}

// WithSonicSerializer uses bytedance/sonic as json serializer
func WithSonicSerializer() Option {
	return func(s *Server) {
		s.e.JSONSerializer = &sonicJSONSerializer{sonic.Config{
			CopyString:       true,
			CompactMarshaler: true,
			EscapeHTML:       true,
			NoNullSliceOrMap: true,
			ValidateString:   true,
		}.Froze()}
	}
}
