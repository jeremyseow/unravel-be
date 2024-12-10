package parameter

import (
	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/server/handler/parameter"
)

func RegisterRoutes(router *gin.Engine, parameterHandler *parameter.ParameterHandler) {
	schemaGroup := router.Group("/parameters")
	{
		schemaGroup.GET("/", parameterHandler.GetParameters)
	}
}
