package schema

import (
	"net/http"

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
		c.Error(err)
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

	created, err := h.SchemaService.CreateSchema(c.Request.Context(), schema)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *SchemaHandler) ListSchemas(c *gin.Context) {
	filter := domain.ListSchemasFilter{}
	if lc := c.Query("lifecycle"); lc != "" {
		filter.Lifecycle = &lc
	}
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	schemas, err := h.SchemaService.ListSchemas(c.Request.Context(), filter)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, schemas)
}

func (h *SchemaHandler) GetSchemas(c *gin.Context) {
	key := c.Param("key")
	schemas, err := h.SchemaService.GetSchemas(c.Request.Context(), key)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, schemas)
}

func (h *SchemaHandler) GetSchemaVersion(c *gin.Context) {
	key := c.Param("key")
	version := c.Param("version")

	schema, err := h.SchemaService.GetSchemaVersion(c.Request.Context(), key, version)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, schema)
}

func (h *SchemaHandler) PublishSchema(c *gin.Context) {
	key := c.Param("key")

	published, err := h.SchemaService.PublishSchema(c.Request.Context(), key)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, published)
}

func (h *SchemaHandler) DeprecateSchema(c *gin.Context) {
	key := c.Param("key")
	version := c.Param("version")

	deprecated, err := h.SchemaService.DeprecateSchema(c.Request.Context(), key, version)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, deprecated)
}
