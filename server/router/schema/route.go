package schema

import (
	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/server/handler/schema"
)

// RegisterRoutes registers all schema-related routes
func RegisterRoutes(router *gin.Engine, schemaHandler *schema.SchemaHandler) {
	schemaGroup := router.Group("/schemas")
	{
		// Create a new schema
		schemaGroup.POST("", schemaHandler.CreateSchema)

		// Get all versions of a schema
		schemaGroup.GET("/:name", schemaHandler.GetSchemas)

		// Get a specific version of a schema
		schemaGroup.GET("/:name/versions/:version", schemaHandler.GetSchemaVersion)

		// Delete a specific version (draft lifecycle only)
		schemaGroup.DELETE("/:name/versions/:version", schemaHandler.DeleteSchemaVersion)

		// Promote the current draft to active and assign a semver
		schemaGroup.POST("/:name/draft/publish", schemaHandler.PublishDraft)

		// Deprecate a specific active version
		schemaGroup.POST("/:name/versions/:version/deprecate", schemaHandler.DeprecateSchemaVersion)
	}
}
