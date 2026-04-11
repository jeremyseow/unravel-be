package parameter

import (
	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/server/handler/parameter"
)

func RegisterRoutes(router *gin.Engine, parameterHandler *parameter.ParameterHandler) {
	group := router.Group("/parameters")
	{
		group.GET("", parameterHandler.GetParameters)
		group.POST("", parameterHandler.CreateParameter)
		group.PUT("/:key", parameterHandler.UpdateParameter)
		group.DELETE("/:key", parameterHandler.DeleteParameter)
	}
}
