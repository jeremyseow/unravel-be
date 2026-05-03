package parameter

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, h *ParameterHandler) {
	parameterGroup := router.Group("/parameters")
	{
		parameterGroup.GET("", h.GetParameters)
		parameterGroup.POST("", h.CreateParameter)
		parameterGroup.PUT("/:key", h.UpdateParameter)
		parameterGroup.DELETE("/:key", h.DeleteParameter)
	}
}
