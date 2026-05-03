package schema

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, schemaHandler *SchemaHandler) {
	schemaGroup := router.Group("/schemas")
	{
		schemaGroup.POST("", schemaHandler.CreateSchema)
		schemaGroup.GET("/:key", schemaHandler.GetSchemas)
		schemaGroup.GET("/:key/versions/:version", schemaHandler.GetSchemaVersion)
		schemaGroup.POST("/:key/publish", schemaHandler.PublishSchema)
		schemaGroup.POST("/:key/versions/:version/deprecate", schemaHandler.DeprecateSchema)
	}
}
