package registry

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, h *RegistryHandler) {
	g := router.Group("/registry")
	{
		g.GET("/:key/latest", h.GetLatestSchema)
		g.GET("/:key/versions/:version", h.GetSchemaVersion)
	}
}
