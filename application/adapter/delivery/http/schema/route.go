package schema

import (
	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/adapter/delivery/http/apierror"
)

func RegisterRoutes(router *gin.RouterGroup, schemaHandler *SchemaHandler) {
	schemaGroup := router.Group("/schemas")
	{
		schemaGroup.POST("", apierror.Wrap(schemaHandler.CreateSchema))
		schemaGroup.GET("/:key", apierror.Wrap(schemaHandler.GetSchemas))
		schemaGroup.GET("/:key/versions/:version", apierror.Wrap(schemaHandler.GetSchemaVersion))
	}
}
