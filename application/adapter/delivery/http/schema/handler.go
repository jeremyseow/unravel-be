package schema

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/application/usecase"
)

type SchemaHandler struct {
	SchemaService usecase.SchemaService
}

func NewSchemaHandler(schemaService usecase.SchemaService) *SchemaHandler {
	return &SchemaHandler{SchemaService: schemaService}
}

func (h *SchemaHandler) CreateSchema(c *gin.Context) {
	var req SchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := make([]domain.SchemaParameter, len(req.Parameters))
	for i, p := range req.Parameters {
		params[i] = domain.SchemaParameter{
			ParameterKey: p.ParameterKey,
			IsRequired:   p.IsRequired,
		}
	}

	schema := domain.Schema{
		SchemaKey:   req.SchemaKey,
		SchemaName:  req.SchemaName,
		Description: req.Description,
		Parameters:  params,
	}

	created, err := h.SchemaService.CreateSchema(c, schema)
	if err != nil {
		if strings.Contains(err.Error(), "parameter keys not found in catalog") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *SchemaHandler) GetSchemas(c *gin.Context) {
	key := c.Param("key")
	schemas, err := h.SchemaService.GetSchemas(c, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, schemas)
}

func (h *SchemaHandler) GetSchemaVersion(c *gin.Context) {
	key := c.Param("key")
	version := c.Param("version")

	schema, err := h.SchemaService.GetSchemaVersion(c, key, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}
	c.JSON(http.StatusOK, schema)
}
