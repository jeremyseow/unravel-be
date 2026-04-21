package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/parameter"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/schema"
)

func SetupRoutes(httpRouter *gin.Engine, allHandlers *AllHandlers) {
	httpRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	schema.RegisterRoutes(httpRouter, allHandlers.SchemaHandler)
	parameter.RegisterRoutes(httpRouter, allHandlers.ParameterHandler)
}
