package parameter

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, h *ParameterHandler) {
	g := router.Group("/parameters")
	{
		g.GET("", h.GetParameters)
		g.POST("", h.CreateParameter)
		g.PUT("/:key", h.UpdateParameter)
		g.DELETE("/:key", h.DeleteParameter)
	}
}
