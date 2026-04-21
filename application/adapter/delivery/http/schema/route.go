package schema

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all schema-related routes
func RegisterRoutes(router *gin.Engine, schemaHandler *SchemaHandler) {
	schemaGroup := router.Group("/schemas")
	{
		// Create a new schema
		schemaGroup.POST("", schemaHandler.CreateSchema)

		// Get all versions of a schema
		schemaGroup.GET("/:name", schemaHandler.GetSchemas)

		// Get a specific version of a schema
		schemaGroup.GET("/:name/versions/:version", schemaHandler.GetSchemaVersion)
	}
}
