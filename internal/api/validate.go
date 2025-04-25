package api

import (
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
	var errs []ValidationError
	err := validate.Struct(s)
	if err != nil {
		for _, vErr := range err.(validator.ValidationErrors) {
			errs = append(
				errs,
				ValidationError{
					Message: DefaultErrorDetailMessage,
					Code:    DefaultErrorDetailCode,
					Field:   vErr.Field(),
				},
			)
		}
	}
	return errs
}
