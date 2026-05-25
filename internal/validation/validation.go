package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// New создает и настраивает новый экземпляр валидатора.
func New() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// Сначала пытаемся прочитать тег 'json'
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		if name == "" {
			name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		}
		return name
	})

	return validate
}
