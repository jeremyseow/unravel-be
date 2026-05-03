package apierror

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type fieldError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type errorResponse struct {
	Errors []fieldError `json:"errors"`
}

// Validation handles errors from ShouldBindJSON, expanding validator.ValidationErrors
// into per-field messages. Falls back to a single message for non-validation errors.
func Validation(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, errorResponse{Errors: formatValidation(err)})
}

// BadRequest writes a 400 with the error message as-is.
func BadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, errorResponse{Errors: []fieldError{{Message: err.Error()}}})
}

// NotFound writes a 404 with the error message as-is.
func NotFound(c *gin.Context, err error) {
	c.JSON(http.StatusNotFound, errorResponse{Errors: []fieldError{{Message: err.Error()}}})
}

// Internal writes a 500 with a generic message. The actual error is not exposed
// to the caller to avoid leaking implementation details.
func Internal(c *gin.Context, _ error) {
	c.JSON(http.StatusInternalServerError, errorResponse{Errors: []fieldError{{Message: "an internal error occurred"}}})
}

func formatValidation(err error) []fieldError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []fieldError{{Message: err.Error()}}
	}
	out := make([]fieldError, len(ve))
	for i, fe := range ve {
		out[i] = fieldError{
			Field:   toSnakeCase(fe.Field()),
			Message: tagToMessage(fe.Tag(), fe.Param()),
		}
	}
	return out
}

func tagToMessage(tag, param string) string {
	switch tag {
	case "required":
		return "is required"
	case "min":
		return "must be at least " + param + " characters"
	case "max":
		return "must be at most " + param + " characters"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(param, " ", ", ")
	default:
		return "is invalid"
	}
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
