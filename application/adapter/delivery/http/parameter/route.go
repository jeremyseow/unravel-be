package parameter

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, parameterHandler *ParameterHandler) {
	schemaGroup := router.Group("/parameters")
	{
		schemaGroup.GET("", parameterHandler.GetParameters)
	}
}
