package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/middleware"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/parameter"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/schema"
)

func SetupRoutes(httpRouter *gin.Engine, allHandlers *AllHandlers) {
	httpRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	api := httpRouter.Group("/")
	api.Use(middleware.TenantMiddleware())

	schema.RegisterRoutes(api, allHandlers.SchemaHandler)
	parameter.RegisterRoutes(api, allHandlers.ParameterHandler)
}
