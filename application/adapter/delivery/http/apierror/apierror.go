package apierror

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jeremyseow/unravel-be/application/domain"
)

// HandlerFunc is a gin handler that returns an error. Use Wrap to register it.
type HandlerFunc func(*gin.Context) error

// Wrap converts a HandlerFunc into a gin.HandlerFunc, routing any returned
// error through Handle.
func Wrap(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h(c); err != nil {
			Handle(c, err)
		}
	}
}

// ValidationError wraps err so Handle treats it as a 400 validation response.
func ValidationError(err error) error {
	return &validationErr{err: err}
}

type validationErr struct{ err error }

func (e *validationErr) Error() string { return e.err.Error() }
func (e *validationErr) Unwrap() error { return e.err }

type fieldError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type errorResponse struct {
	Errors []fieldError `json:"errors"`
}

// Handle maps a domain error to the appropriate HTTP response. Add new sentinel
// errors here as the domain grows; handlers never need to change.
func Handle(c *gin.Context, err error) {
	var ve *validationErr
	switch {
	case errors.As(err, &ve):
		Validation(c, ve.err)
	case errors.Is(err, domain.ErrNotFound):
		NotFound(c, err)
	case errors.Is(err, domain.ErrInvalidDataType),
		errors.Is(err, domain.ErrParameterKeysNotFound):
		BadRequest(c, err)
	default:
		Internal(c, err)
	}
}

// Validation writes a 400 with per-field detail from a validator.ValidationErrors.
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
