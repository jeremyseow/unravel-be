package schema

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	schemaStorage "github.com/jeremyseow/unravel-be/storage/schema"
)

type SchemaHandler struct {
	SchemaStorage schemaStorage.Storage
}

func NewSchemaHandler(storage schemaStorage.Storage) *SchemaHandler {
	return &SchemaHandler{SchemaStorage: storage}
}

// CreateSchema handles POST /schemas.
// All new schemas start as drafts at version 0.0.0. The real semver is
// assigned at publish time based on what changed from the previous version.
func (h *SchemaHandler) CreateSchema(c *gin.Context) {
	var req CreateSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isLatest := false
	lifecycle := "draft"

	schemaRow := model.EntitySchemas{
		SchemaKey:     req.SchemaKey,
		SchemaName:    req.SchemaName,
		SchemaVersion: schemaStorage.DraftVersion,
		Description:   req.Description,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}

	mappings := make([]model.EntitySchemasParametersMappings, 0, len(req.Parameters))
	for _, p := range req.Parameters {
		isRequired := p.IsRequired
		mappings = append(mappings, model.EntitySchemasParametersMappings{
			ParameterKey: p.ParameterKey,
			IsRequired:   &isRequired,
		})
	}

	record, err := h.SchemaStorage.CreateSchema(c, schemaRow, mappings)
	if err != nil {
		if strings.Contains(err.Error(), "parameter keys not found in catalog") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toSchemaResponse(record))
}

// GetSchemas handles GET /schemas/:name — returns all versions of a schema.
func (h *SchemaHandler) GetSchemas(c *gin.Context) {
	key := c.Param("name")

	records, err := h.SchemaStorage.GetSchemasByKey(c, key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	responses := make([]SchemaResponse, 0, len(records))
	for i := range records {
		responses = append(responses, toSchemaResponse(&records[i]))
	}
	c.JSON(http.StatusOK, responses)
}

// GetSchemaVersion handles GET /schemas/:name/versions/:version.
func (h *SchemaHandler) GetSchemaVersion(c *gin.Context) {
	key := c.Param("name")
	version := c.Param("version")

	record, err := h.SchemaStorage.GetSchemaVersion(c, key, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toSchemaResponse(record))
}

// DeleteSchemaVersion handles DELETE /schemas/:name/versions/:version.
// Only draft schemas may be deleted; active schemas are immutable.
func (h *SchemaHandler) DeleteSchemaVersion(c *gin.Context) {
	key := c.Param("name")
	version := c.Param("version")

	if err := h.SchemaStorage.DeleteSchemaVersion(c, key, version); err != nil {
		if strings.HasPrefix(err.Error(), "schema_not_draft:") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// PublishDraft handles POST /schemas/:name/draft/publish.
// Promotes the current draft to active, auto-detects the semver bump, and
// assigns the new version. Returns the published schema with bump metadata.
func (h *SchemaHandler) PublishDraft(c *gin.Context) {
	key := c.Param("name")

	result, err := h.SchemaStorage.PublishDraft(c, key)
	if err != nil {
		if strings.HasPrefix(err.Error(), "no draft found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toPublishResponse(result))
}

// DeprecateSchemaVersion handles POST /schemas/:name/versions/:version/deprecate.
func (h *SchemaHandler) DeprecateSchemaVersion(c *gin.Context) {
	key := c.Param("name")
	version := c.Param("version")

	record, err := h.SchemaStorage.DeprecateSchemaVersion(c, key, version)
	if err != nil {
		if strings.HasPrefix(err.Error(), "schema_not_active:") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toSchemaResponse(record))
}

func toPublishResponse(r *schemaStorage.PublishResult) PublishResponse {
	return PublishResponse{
		SchemaResponse:  toSchemaResponse(&r.SchemaRecord),
		VersionBump:     r.VersionBump,
		PreviousVersion: r.PreviousVersion,
	}
}

func toSchemaResponse(r *schemaStorage.SchemaRecord) SchemaResponse {
	params := make([]SchemaParamDetail, 0, len(r.Parameters))
	for _, m := range r.Parameters {
		isRequired := false
		if m.IsRequired != nil {
			isRequired = *m.IsRequired
		}
		params = append(params, SchemaParamDetail{
			ParameterKey: m.ParameterKey,
			IsRequired:   isRequired,
		})
	}

	return SchemaResponse{
		ID:          r.ID,
		SchemaKey:   r.SchemaKey,
		SchemaName:  r.SchemaName,
		Version:     r.SchemaVersion,
		Description: r.Description,
		IsLatest:    r.IsLatest,
		Lifecycle:   r.Lifecycle,
		Parameters:  params,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
