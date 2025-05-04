package tools

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const ValidateRulesTagName = "v"

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
	validate.SetTagName(ValidateRulesTagName)
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		jsonTag := field.Tag.Get("json")
		if len(jsonTag) == 0 || jsonTag == "-" {
			return ""
		}
		return strings.
			ReplaceAll(jsonTag, ",omitempty", "")
	})
}

func ValidateStruct(s interface{}) []ValidationError {
	err := validate.Struct(s)
	if err != nil {
		//goland:noinspection GoTypeAssertionOnErrors
		switch e := err.(type) {
		case validator.ValidationErrors:
			vErrs := make([]ValidationError, 0, len(e))
			for _, vErr := range e {
				vErrs = append(vErrs, ValidationError{original: vErr})
			}
			return vErrs
		default:
			slog.Error("validate.ValidateStruct error", slog.Any("error", e))
			return []ValidationError{{
				pointer: "UNKNOWN",
				detail:  e.Error(),
				code:    "INTERNAL_VALIDATION_ERROR",
			}}
		}
	}
	return nil
}

// ---

type ValidationError struct {
	original validator.FieldError
	pointer  string
	detail   string
	code     string
}

func (e ValidationError) Pointer() string {
	if e.pointer != "" {
		return e.pointer
	}
	return "#/" + strings.ReplaceAll(e.original.Namespace(), ".", "/")
}

func (e ValidationError) Detail() string {
	if e.detail != "" {
		return e.detail
	}
	return fmt.Sprintf("rule `%s` with value of %s", e.original.ActualTag(), e.original.Param())
}

func (e ValidationError) Code() string {
	if e.code != "" {
		return e.code
	}
	return "INVALID_VALUE"
}

func (e ValidationError) Compact() string {
	return fmt.Sprintf("(%s)[%s='%s']", e.original.Field(), e.original.ActualTag(), e.original.Param())
}
