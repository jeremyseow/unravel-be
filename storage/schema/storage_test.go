package schema

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

// openTestDB opens a PostgreSQL connection from the TEST_DB_DSN environment
// variable. If the variable is absent the test is skipped — this keeps
// unit-only CI runs (no DB) from failing.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set — skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

// seedParameter inserts a minimal parameter row for the default tenant and
// registers a cleanup function that removes it after the test.
func seedParameter(t *testing.T, db *sql.DB, key string) {
	t.Helper()
	stmt := EntityParameters.INSERT(
		EntityParameters.TenantID,
		EntityParameters.ParameterKey,
		EntityParameters.ParameterName,
		EntityParameters.DataType,
	).VALUES(
		defaultTenantID,
		key,
		key+"_name",
		"string",
	)
	_, err := stmt.Exec(db)
	require.NoError(t, err, "seed parameter %s", key)

	t.Cleanup(func() {
		EntityParameters.DELETE().WHERE(
			EntityParameters.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntityParameters.ParameterKey.EQ(String(key))),
		).Exec(db) //nolint:errcheck — best-effort cleanup
	})
}

// cleanupSchema removes all schema rows (and their mappings via FK cascade)
// for a given schema key after the test.
func cleanupSchema(t *testing.T, db *sql.DB, key string) {
	t.Helper()
	t.Cleanup(func() {
		EntitySchemasParametersMappings.DELETE().WHERE(
			EntitySchemasParametersMappings.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(key))),
		).Exec(db) //nolint:errcheck
		EntitySchemas.DELETE().WHERE(
			EntitySchemas.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntitySchemas.SchemaKey.EQ(String(key))),
		).Exec(db) //nolint:errcheck
	})
}

// uniqueKey generates a test-scoped schema key that won't collide across runs.
func uniqueKey(t *testing.T) string {
	t.Helper()
	// Truncate test name and append a nanosecond suffix to stay within VARCHAR(32).
	name := strings.ToLower(t.Name())
	name = strings.NewReplacer("/", "_", " ", "_").Replace(name)
	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1e9)
	key := fmt.Sprintf("t%s", suffix) // always starts with letter
	if len(key) > 32 {
		key = key[:32]
	}
	_ = name // kept for readability in failure messages
	return key
}

// --- CreateSchema ---

func TestCreateSchema_Success(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	storage := NewStorage(db)

	paramKey := uniqueKey(t) + "p"
	schemaKey := uniqueKey(t) + "s"
	seedParameter(t, db, paramKey)
	cleanupSchema(t, db, schemaKey)

	isRequired := true
	mappings := []model.EntitySchemasParametersMappings{
		{ParameterKey: paramKey, IsRequired: &isRequired},
	}
	isLatest := false
	lifecycle := "draft"
	schema := model.EntitySchemas{
		SchemaKey:     schemaKey,
		SchemaName:    "Test Schema",
		SchemaVersion: DraftVersion,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}

	record, err := storage.CreateSchema(context.Background(), schema, mappings)
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, schemaKey, record.SchemaKey)
	assert.Equal(t, DraftVersion, record.SchemaVersion)
	assert.Equal(t, "draft", *record.Lifecycle)
	assert.False(t, *record.IsLatest)
	require.Len(t, record.Parameters, 1)
	assert.Equal(t, paramKey, record.Parameters[0].ParameterKey)
	assert.True(t, *record.Parameters[0].IsRequired)
}

func TestCreateSchema_UnknownParameterKey(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	storage := NewStorage(db)

	schemaKey := uniqueKey(t) + "s"
	cleanupSchema(t, db, schemaKey)

	isRequired := false
	mappings := []model.EntitySchemasParametersMappings{
		{ParameterKey: "does_not_exist_xyz", IsRequired: &isRequired},
	}
	isLatest := false
	lifecycle := "draft"
	schema := model.EntitySchemas{
		SchemaKey:     schemaKey,
		SchemaName:    "Bad Schema",
		SchemaVersion: DraftVersion,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}

	_, err := storage.CreateSchema(context.Background(), schema, mappings)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parameter keys not found in catalog")
}

func TestCreateSchema_DuplicateDraft(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	storage := NewStorage(db)

	paramKey := uniqueKey(t) + "p"
	schemaKey := uniqueKey(t) + "s"
	seedParameter(t, db, paramKey)
	cleanupSchema(t, db, schemaKey)

	isRequired := false
	mappings := []model.EntitySchemasParametersMappings{
		{ParameterKey: paramKey, IsRequired: &isRequired},
	}
	isLatest := false
	lifecycle := "draft"
	schema := model.EntitySchemas{
		SchemaKey:     schemaKey,
		SchemaName:    "Dup Schema",
		SchemaVersion: DraftVersion,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}

	_, err := storage.CreateSchema(context.Background(), schema, mappings)
	require.NoError(t, err, "first create should succeed")

	_, err = storage.CreateSchema(context.Background(), schema, mappings)
	require.Error(t, err, "second create should violate the single-draft unique index")
}

// --- DeleteSchemaVersion ---

func TestDeleteSchemaVersion_Draft(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	storage := NewStorage(db)

	paramKey := uniqueKey(t) + "p"
	schemaKey := uniqueKey(t) + "s"
	seedParameter(t, db, paramKey)
	cleanupSchema(t, db, schemaKey) // handles the case where delete fails

	isRequired := false
	mappings := []model.EntitySchemasParametersMappings{
		{ParameterKey: paramKey, IsRequired: &isRequired},
	}
	isLatest := false
	lifecycle := "draft"
	_, err := storage.CreateSchema(context.Background(), model.EntitySchemas{
		SchemaKey:     schemaKey,
		SchemaName:    "Del Schema",
		SchemaVersion: DraftVersion,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}, mappings)
	require.NoError(t, err)

	err = storage.DeleteSchemaVersion(context.Background(), schemaKey, DraftVersion)
	require.NoError(t, err)

	// Confirm it's gone.
	_, err = storage.GetSchemaVersion(context.Background(), schemaKey, DraftVersion)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema version not found")
}

func TestDeleteSchemaVersion_ActiveIsRejected(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	storage := NewStorage(db)

	paramKey := uniqueKey(t) + "p"
	schemaKey := uniqueKey(t) + "s"
	seedParameter(t, db, paramKey)
	cleanupSchema(t, db, schemaKey)

	isRequired := false
	mappings := []model.EntitySchemasParametersMappings{
		{ParameterKey: paramKey, IsRequired: &isRequired},
	}
	isLatest := false
	lifecycle := "draft"
	_, err := storage.CreateSchema(context.Background(), model.EntitySchemas{
		SchemaKey:     schemaKey,
		SchemaName:    "Active Schema",
		SchemaVersion: DraftVersion,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}, mappings)
	require.NoError(t, err)

	_, err = storage.PublishDraft(context.Background(), schemaKey)
	require.NoError(t, err)

	err = storage.DeleteSchemaVersion(context.Background(), schemaKey, "1.0.0")
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "schema_not_draft:"), "error should have schema_not_draft: prefix, got: %s", err.Error())
}

// --- PublishDraft ---

func TestPublishDraft_FirstPublish(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	storage := NewStorage(db)

	paramKey := uniqueKey(t) + "p"
	schemaKey := uniqueKey(t) + "s"
	seedParameter(t, db, paramKey)
	cleanupSchema(t, db, schemaKey)

	isRequired := true
	isLatest := false
	lifecycle := "draft"
	_, err := storage.CreateSchema(context.Background(), model.EntitySchemas{
		SchemaKey:     schemaKey,
		SchemaName:    "First Pub",
		SchemaVersion: DraftVersion,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}, []model.EntitySchemasParametersMappings{
		{ParameterKey: paramKey, IsRequired: &isRequired},
	})
	require.NoError(t, err)

	result, err := storage.PublishDraft(context.Background(), schemaKey)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "1.0.0", result.SchemaVersion)
	assert.Equal(t, "major", result.VersionBump)
	assert.Empty(t, result.PreviousVersion)
	assert.Equal(t, "active", *result.Lifecycle)
	assert.True(t, *result.IsLatest)
}

func TestPublishDraft_NoDraft(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	storage := NewStorage(db)

	_, err := storage.PublishDraft(context.Background(), "does_not_exist_key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no draft found")
}
