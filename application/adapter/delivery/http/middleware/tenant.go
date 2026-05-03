package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/ctxkey"
)

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := ctxkey.WithTenantID(c.Request.Context(), ctxkey.DefaultTenantID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
