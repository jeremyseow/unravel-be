package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	schemaStorage "github.com/jeremyseow/unravel-be/storage/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorage is a hand-crafted implementation of schemaStorage.Storage for tests.
// Each method field can be overridden per test case.
type mockStorage struct {
	createSchema          func(ctx context.Context, schema model.EntitySchemas, mappings []model.EntitySchemasParametersMappings) (*schemaStorage.SchemaRecord, error)
	getSchemasByKey       func(ctx context.Context, key string) ([]schemaStorage.SchemaRecord, error)
	getSchemaVersion      func(ctx context.Context, key string, version string) (*schemaStorage.SchemaRecord, error)
	deleteSchemaVersion   func(ctx context.Context, key string, version string) error
	publishDraft          func(ctx context.Context, key string) (*schemaStorage.PublishResult, error)
	deprecateSchemaVersion func(ctx context.Context, key string, version string) (*schemaStorage.SchemaRecord, error)
}

func (m *mockStorage) CreateSchema(ctx context.Context, schema model.EntitySchemas, mappings []model.EntitySchemasParametersMappings) (*schemaStorage.SchemaRecord, error) {
	return m.createSchema(ctx, schema, mappings)
}
func (m *mockStorage) GetSchemasByKey(ctx context.Context, key string) ([]schemaStorage.SchemaRecord, error) {
	return m.getSchemasByKey(ctx, key)
}
func (m *mockStorage) GetSchemaVersion(ctx context.Context, key string, version string) (*schemaStorage.SchemaRecord, error) {
	return m.getSchemaVersion(ctx, key, version)
}
func (m *mockStorage) DeleteSchemaVersion(ctx context.Context, key string, version string) error {
	return m.deleteSchemaVersion(ctx, key, version)
}
func (m *mockStorage) PublishDraft(ctx context.Context, key string) (*schemaStorage.PublishResult, error) {
	return m.publishDraft(ctx, key)
}
func (m *mockStorage) DeprecateSchemaVersion(ctx context.Context, key string, version string) (*schemaStorage.SchemaRecord, error) {
	return m.deprecateSchemaVersion(ctx, key, version)
}

// newTestRouter wires up a Gin router in test mode with the given handler.
func newTestRouter(h *SchemaHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/schemas", h.CreateSchema)
	r.GET("/schemas/:name", h.GetSchemas)
	r.GET("/schemas/:name/versions/:version", h.GetSchemaVersion)
	r.DELETE("/schemas/:name/versions/:version", h.DeleteSchemaVersion)
	r.POST("/schemas/:name/draft/publish", h.PublishDraft)
	r.POST("/schemas/:name/versions/:version/deprecate", h.DeprecateSchemaVersion)
	return r
}

// sampleRecord returns a minimal SchemaRecord for use as a mock return value.
func sampleRecord() *schemaStorage.SchemaRecord {
	isLatest := false
	lifecycle := "draft"
	isRequired := true
	return &schemaStorage.SchemaRecord{
		EntitySchemas: model.EntitySchemas{
			ID:            1,
			SchemaKey:     "test_event",
			SchemaName:    "Test Event",
			SchemaVersion: schemaStorage.DraftVersion,
			IsLatest:      &isLatest,
			Lifecycle:     &lifecycle,
		},
		Parameters: []model.EntitySchemasParametersMappings{
			{ParameterKey: "user_id", IsRequired: &isRequired},
		},
	}
}

// --- CreateSchema tests ---

func TestCreateSchema_Success(t *testing.T) {
	storage := &mockStorage{
		createSchema: func(_ context.Context, _ model.EntitySchemas, _ []model.EntitySchemasParametersMappings) (*schemaStorage.SchemaRecord, error) {
			return sampleRecord(), nil
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	body := `{
		"schema_key": "test_event",
		"schema_name": "Test Event",
		"parameters": [{"parameter_key": "user_id", "is_required": true}]
	}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp SchemaResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "test_event", resp.SchemaKey)
	assert.Equal(t, schemaStorage.DraftVersion, resp.Version)
	assert.Len(t, resp.Parameters, 1)
	assert.Equal(t, "user_id", resp.Parameters[0].ParameterKey)
	assert.True(t, resp.Parameters[0].IsRequired)
}

func TestCreateSchema_MissingSchemaKey(t *testing.T) {
	router := newTestRouter(NewSchemaHandler(&mockStorage{}))

	body := `{"schema_name": "Test Event", "parameters": [{"parameter_key": "user_id"}]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchema_MissingSchemaName(t *testing.T) {
	router := newTestRouter(NewSchemaHandler(&mockStorage{}))

	body := `{"schema_key": "test_event", "parameters": [{"parameter_key": "user_id"}]}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchema_EmptyParameters(t *testing.T) {
	router := newTestRouter(NewSchemaHandler(&mockStorage{}))

	body := `{"schema_key": "test_event", "schema_name": "Test Event", "parameters": []}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchema_MissingParameters(t *testing.T) {
	router := newTestRouter(NewSchemaHandler(&mockStorage{}))

	body := `{"schema_key": "test_event", "schema_name": "Test Event"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchema_UnknownParameterKey(t *testing.T) {
	storage := &mockStorage{
		createSchema: func(_ context.Context, _ model.EntitySchemas, _ []model.EntitySchemasParametersMappings) (*schemaStorage.SchemaRecord, error) {
			return nil, fmt.Errorf("parameter keys not found in catalog: [ghost_param]")
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	body := `{
		"schema_key": "test_event",
		"schema_name": "Test Event",
		"parameters": [{"parameter_key": "ghost_param"}]
	}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSchema_StorageError(t *testing.T) {
	storage := &mockStorage{
		createSchema: func(_ context.Context, _ model.EntitySchemas, _ []model.EntitySchemasParametersMappings) (*schemaStorage.SchemaRecord, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	body := `{
		"schema_key": "test_event",
		"schema_name": "Test Event",
		"parameters": [{"parameter_key": "user_id"}]
	}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- PublishDraft tests ---

func TestPublishDraft_Success(t *testing.T) {
	isLatest := true
	lifecycle := "active"
	bump := "major"
	isRequired := false
	storage := &mockStorage{
		publishDraft: func(_ context.Context, key string) (*schemaStorage.PublishResult, error) {
			assert.Equal(t, "test_event", key)
			return &schemaStorage.PublishResult{
				SchemaRecord: schemaStorage.SchemaRecord{
					EntitySchemas: model.EntitySchemas{
						SchemaKey:     "test_event",
						SchemaName:    "Test Event",
						SchemaVersion: "1.0.0",
						IsLatest:      &isLatest,
						Lifecycle:     &lifecycle,
					},
					Parameters: []model.EntitySchemasParametersMappings{
						{ParameterKey: "user_id", IsRequired: &isRequired},
					},
				},
				VersionBump:     bump,
				PreviousVersion: "",
			}, nil
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas/test_event/draft/publish", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp PublishResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "1.0.0", resp.Version)
	assert.Equal(t, "major", resp.VersionBump)
	assert.Empty(t, resp.PreviousVersion)
}

func TestPublishDraft_NoDraftFound(t *testing.T) {
	storage := &mockStorage{
		publishDraft: func(_ context.Context, key string) (*schemaStorage.PublishResult, error) {
			return nil, fmt.Errorf("no draft found for schema %s", key)
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas/test_event/draft/publish", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- DeprecateSchemaVersion tests ---

func TestDeprecateSchemaVersion_Success(t *testing.T) {
	isLatest := false
	lifecycle := "deprecated"
	storage := &mockStorage{
		deprecateSchemaVersion: func(_ context.Context, key, version string) (*schemaStorage.SchemaRecord, error) {
			assert.Equal(t, "test_event", key)
			assert.Equal(t, "1.0.0", version)
			return &schemaStorage.SchemaRecord{
				EntitySchemas: model.EntitySchemas{
					SchemaKey:     key,
					SchemaVersion: version,
					IsLatest:      &isLatest,
					Lifecycle:     &lifecycle,
				},
			}, nil
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas/test_event/versions/1.0.0/deprecate", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp SchemaResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "deprecated", *resp.Lifecycle)
}

func TestDeprecateSchemaVersion_NotActive(t *testing.T) {
	storage := &mockStorage{
		deprecateSchemaVersion: func(_ context.Context, key, version string) (*schemaStorage.SchemaRecord, error) {
			return nil, fmt.Errorf("schema_not_active: schema %s@%s has lifecycle \"draft\"", key, version)
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schemas/test_event/versions/0.0.0/deprecate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- DeleteSchemaVersion tests ---

func TestDeleteSchemaVersion_Success(t *testing.T) {
	storage := &mockStorage{
		deleteSchemaVersion: func(_ context.Context, key, version string) error {
			return nil
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/schemas/test_event/versions/0.0.0", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteSchemaVersion_NotDraft(t *testing.T) {
	storage := &mockStorage{
		deleteSchemaVersion: func(_ context.Context, key, version string) error {
			return fmt.Errorf("schema_not_draft: schema %s@%s has lifecycle \"active\"; only draft schemas can be deleted", key, version)
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/schemas/test_event/versions/1.0.0", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDeleteSchemaVersion_NotFound(t *testing.T) {
	storage := &mockStorage{
		deleteSchemaVersion: func(_ context.Context, key, version string) error {
			return fmt.Errorf("schema version not found: %s@%s", key, version)
		},
	}
	router := newTestRouter(NewSchemaHandler(storage))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/schemas/test_event/versions/9.9.9", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
