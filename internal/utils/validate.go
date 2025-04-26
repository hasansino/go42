package utils

import (
	"fmt"
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

type ValidationError struct {
	Field string `json:"message"`
	Rule  string `json:"rule"`
}

type ValidationErrors []ValidationError

func (vErrs ValidationErrors) Strings() []string {
	s := make([]string, 0, len(vErrs))
	for _, vErr := range vErrs {
		s = append(s, fmt.Sprintf("(%s)[%s]", vErr.Field, vErr.Rule))
	}
	return s
}

func ValidateStruct(s interface{}) ValidationErrors {
	var errs []ValidationError
	err := validate.Struct(s)
	if err != nil {
		for _, vErr := range err.(validator.ValidationErrors) {
			errs = append(
				errs,
				ValidationError{
					Field: vErr.Namespace(),
					Rule:  fmt.Sprintf("%s='%s'", vErr.ActualTag(), vErr.Param()),
				},
			)
		}
	}
	return errs
}
