package schema

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/application/usecase"
)

// SchemaHandler handles schema-related operations
type SchemaHandler struct {
	SchemaService usecase.SchemaService
}

// NewSchemaHandler creates a new schema handler
func NewSchemaHandler(schemaService usecase.SchemaService) *SchemaHandler {
	return &SchemaHandler{
		SchemaService: schemaService,
	}
}

// CreateSchema handles the creation of a new schema
func (h *SchemaHandler) CreateSchema(c *gin.Context) {
	var req SchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.ValidateParameters(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	description := ""
	schema := domain.Schema{
		TenantID:    uuid.Nil, // TODO: Get tenant ID from context
		SchemaKey:   req.Name, // TODO: Generate a schema key
		Description: &description,
	}

	createdSchema, err := h.SchemaService.CreateSchema(c, schema, req.Parameters)
	if err != nil {
		if err.Error() == "schema version already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdSchema)
}

// GetSchemas returns all versions of a schema
func (h *SchemaHandler) GetSchemas(c *gin.Context) {
	name := c.Param("name")
	schemas, err := h.SchemaService.GetSchemas(c, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schemas)
}

// GetSchemaVersion returns a specific version of a schema
func (h *SchemaHandler) GetSchemaVersion(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	schema, err := h.SchemaService.GetSchemaVersion(c, name, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	c.JSON(http.StatusOK, schema)
}

