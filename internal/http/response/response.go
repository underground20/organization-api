package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Message string `json:"message"`
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errors []string
	for _, err := range errs {
		errors = append(errors, fmt.Sprintf(
			"field '%s' - %s",
			strings.ToLower(err.Field()),
			err.Tag(),
		))
	}

	return Response{
		Message: "Validation error: " + strings.Join(errors, "; "),
	}
}

func UnhandledError() Response {
	return Response{Message: "Unhandled error"}
}
