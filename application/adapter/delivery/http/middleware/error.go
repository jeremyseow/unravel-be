package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/samber/lo"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var ve validator.ValidationErrors

			switch {
			case errors.As(err, &ve):
				out := make(map[string]string)
				for _, fe := range ve {
					out[lo.SnakeCase(fe.Field())] = getValidationErrorMsg(fe)
				}
				c.JSON(http.StatusBadRequest, gin.H{"errors": out})
				return
			case errors.Is(err, domain.ErrNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			case errors.Is(err, domain.ErrInvalidDataType):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, domain.ErrParameterKeysNotFound):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			case errors.Is(err, domain.ErrSchemaNotDraft),
				errors.Is(err, domain.ErrSchemaNotActive):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}
}

// getValidationErrorMsg returns a user-friendly error message for a gin validation error.
func getValidationErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "max":
		return fmt.Sprintf("Maximum value is %s", fe.Param())
	case "min":
		return fmt.Sprintf("Minimum value is %s", fe.Param())
	case "url":
		return "Invalid URL format"
	case "alpha_num":
		return "Only letters and numbers are allowed"
	case "email":
		return "Invalid email address"
	default:
		return "Invalid input"
	}
}
