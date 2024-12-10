package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SchemaHandler handles schema-related operations
type SchemaHandler struct {
	// TODO: Add database interface here
	schemas map[string][]Schema // temporary in-memory storage
}

// NewSchemaHandler creates a new schema handler
func NewSchemaHandler() *SchemaHandler {
	return &SchemaHandler{
		schemas: make(map[string][]Schema),
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

	// Generate version based on parameters
	version := h.generateVersion(req.Parameters)

	// Check if this version already exists
	schemas := h.schemas[req.Name]
	for _, s := range schemas {
		if s.Version == version {
			c.JSON(http.StatusConflict, gin.H{"error": "schema version already exists"})
			return
		}
	}

	schema := Schema{
		ID:         uuid.New(),
		Name:       req.Name,
		Version:    version,
		Parameters: req.Parameters,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	h.schemas[req.Name] = append(h.schemas[req.Name], schema)

	c.JSON(http.StatusCreated, schema)
}

// GetSchemas returns all versions of a schema
func (h *SchemaHandler) GetSchemas(c *gin.Context) {
	name := c.Param("name")
	schemas, exists := h.schemas[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}

	c.JSON(http.StatusOK, schemas)
}

// GetSchemaVersion returns a specific version of a schema
func (h *SchemaHandler) GetSchemaVersion(c *gin.Context) {
	name := c.Param("name")
	version := c.Param("version")

	schemas, exists := h.schemas[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
		return
	}

	for _, schema := range schemas {
		if schema.Version == version {
			c.JSON(http.StatusOK, schema)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
}

// generateVersion creates a semantic version based on the parameters
func (h *SchemaHandler) generateVersion(params map[string]any) string {
	// Sort keys for consistent hashing
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create a JSON string of sorted parameters
	paramBytes, _ := json.Marshal(params)

	// Generate hash of parameters
	hash := sha256.Sum256(paramBytes)
	shortHash := hex.EncodeToString(hash[:])[:8]

	// For now, using a simple version scheme: v1-[hash]
	// TODO: Implement proper semantic versioning based on parameter changes
	return "v1-" + shortHash
}
