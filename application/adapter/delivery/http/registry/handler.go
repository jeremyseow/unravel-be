package registry

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/application/usecase"
)

type RegistryHandler struct {
	RegistryService usecase.RegistryService
}

func NewRegistryHandler(svc usecase.RegistryService) *RegistryHandler {
	return &RegistryHandler{RegistryService: svc}
}

func (h *RegistryHandler) GetLatestSchema(c *gin.Context) {
	key := c.Param("key")
	schema, params, err := h.RegistryService.GetLatestSchema(c.Request.Context(), key)
	if err != nil {
		c.Error(err)
		return
	}
	h.respond(c, schema.SchemaKey, params)
}

func (h *RegistryHandler) GetSchemaVersion(c *gin.Context) {
	key := c.Param("key")
	version := c.Param("version")
	schema, params, err := h.RegistryService.GetSchemaVersion(c.Request.Context(), key, version)
	if err != nil {
		c.Error(err)
		return
	}
	h.respond(c, schema.SchemaKey, params)
}

func (h *RegistryHandler) respond(c *gin.Context, key string, params []domain.EnrichedSchemaParameter) {
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "application/x-protobuf") {
		data, err := usecase.GenerateFileDescriptor(key, params)
		if err != nil {
			c.Error(err)
			return
		}
		c.Data(http.StatusOK, "application/x-protobuf", data)
		return
	}
	text := usecase.GenerateProtoText(key, params)
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(text))
}
