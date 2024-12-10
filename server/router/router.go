package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/server/handler"
	"github.com/jeremyseow/unravel-be/server/router/parameter"
	"github.com/jeremyseow/unravel-be/server/router/schema"
)

func SetupRoutes(httpRouter *gin.Engine, allHandlers *handler.AllHandlers) {
	httpRouter.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	schema.RegisterRoutes(httpRouter, allHandlers.SchemaHandler)
	parameter.RegisterRoutes(httpRouter, allHandlers.ParameterHandler)
}
