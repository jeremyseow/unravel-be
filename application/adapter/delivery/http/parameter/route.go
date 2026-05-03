package parameter

import (
	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/apierror"
)

func RegisterRoutes(router *gin.RouterGroup, h *ParameterHandler) {
	g := router.Group("/parameters")
	{
		g.GET("", apierror.Wrap(h.GetParameters))
		g.POST("", apierror.Wrap(h.CreateParameter))
		g.PUT("/:key", apierror.Wrap(h.UpdateParameter))
		g.DELETE("/:key", apierror.Wrap(h.DeleteParameter))
	}
}
