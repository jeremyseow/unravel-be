package schema

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

// testDB is the package-level database connection shared by all integration
// tests. It is initialised in TestMain. A nil value means Docker was not
// available; each test skips itself in that case via requireDB.
var testDB *sql.DB

// TestMain starts a single PostgreSQL container for the entire test binary,
// applies all migrations, and tears the container down after the run.
// Unit tests in this package (versioning_test.go) are unaffected because
// they never call requireDB.
func TestMain(m *testing.M) {
	ctx := context.Background()
	var terminate func()

	pgc, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testcontainers: failed to start postgres container: %v\n", err)
	} else {
		terminate = func() { pgc.Terminate(ctx) } //nolint:errcheck

		connStr, connErr := pgc.ConnectionString(ctx, "sslmode=disable")
		if connErr != nil {
			fmt.Fprintf(os.Stderr, "testcontainers: failed to get connection string: %v\n", connErr)
		} else {
			db, openErr := sql.Open("postgres", connStr)
			if openErr != nil {
				fmt.Fprintf(os.Stderr, "testcontainers: failed to open db: %v\n", openErr)
			} else if migrErr := applyMigrations(db); migrErr != nil {
				fmt.Fprintf(os.Stderr, "testcontainers: failed to apply migrations: %v\n", migrErr)
			} else {
				testDB = db
			}
		}
	}

	code := m.Run()

	if terminate != nil {
		terminate()
	}
	os.Exit(code)
}

// requireDB returns the shared test database or skips the test if the
// PostgreSQL container could not be started (e.g. Docker unavailable in CI).
func requireDB(t *testing.T) *sql.DB {
	t.Helper()
	if testDB == nil {
		t.Skip("PostgreSQL container not available — skipping integration test")
	}
	return testDB
}

// findModuleRoot walks up from the current working directory until it finds
// go.mod, indicating the module root.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

// applyMigrations reads every *.sql file in db/migration/up (sorted by name)
// and executes each one against the provided database. This mirrors the
// behaviour of `make migration_up` without requiring the migrate CLI.
func applyMigrations(db *sql.DB) error {
	root, err := findModuleRoot()
	if err != nil {
		return err
	}
	migDir := filepath.Join(root, "db", "migration", "up")

	entries, err := os.ReadDir(migDir)
	if err != nil {
		return fmt.Errorf("read migration dir %s: %w", migDir, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read %s: %w", entry.Name(), err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// seedParameter inserts a parameter row used as a foreign-key reference in
// schema mapping tests and registers cleanup.
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
		).Exec(db) //nolint:errcheck
	})
}

// cleanupSchema deletes all rows for a schema key after the test, covering
// the case where a test may have failed before its own cleanup ran.
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

// uniqueKey returns a unique schema/parameter key for each test invocation
// to prevent collisions when tests run in parallel or are retried.
func uniqueKey(suffix string) string {
	key := fmt.Sprintf("t%d%s", time.Now().UnixNano()%1e9, suffix)
	if len(key) > 32 {
		key = key[:32]
	}
	return key
}

// newDraftSchema is a helper to build a standard draft EntitySchemas row.
func newDraftSchema(key, name string) model.EntitySchemas {
	isLatest := false
	lifecycle := "draft"
	return model.EntitySchemas{
		SchemaKey:     key,
		SchemaName:    name,
		SchemaVersion: DraftVersion,
		IsLatest:      &isLatest,
		Lifecycle:     &lifecycle,
	}
}

// --- CreateSchema ---

func TestCreateSchema_Success(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	pKey := uniqueKey("p")
	sKey := uniqueKey("s")
	seedParameter(t, db, pKey)
	cleanupSchema(t, db, sKey)

	isRequired := true
	record, err := storage.CreateSchema(context.Background(),
		newDraftSchema(sKey, "Test Schema"),
		[]model.EntitySchemasParametersMappings{
			{ParameterKey: pKey, IsRequired: &isRequired},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, record)

	assert.Equal(t, sKey, record.SchemaKey)
	assert.Equal(t, DraftVersion, record.SchemaVersion)
	assert.Equal(t, "draft", *record.Lifecycle)
	assert.False(t, *record.IsLatest)
	require.Len(t, record.Parameters, 1)
	assert.Equal(t, pKey, record.Parameters[0].ParameterKey)
	assert.True(t, *record.Parameters[0].IsRequired)
}

func TestCreateSchema_UnknownParameterKey(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	sKey := uniqueKey("s")
	cleanupSchema(t, db, sKey)

	isRequired := false
	_, err := storage.CreateSchema(context.Background(),
		newDraftSchema(sKey, "Bad Schema"),
		[]model.EntitySchemasParametersMappings{
			{ParameterKey: "does_not_exist_xyz", IsRequired: &isRequired},
		},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parameter keys not found in catalog")
}

func TestCreateSchema_DuplicateDraft(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	pKey := uniqueKey("p")
	sKey := uniqueKey("s")
	seedParameter(t, db, pKey)
	cleanupSchema(t, db, sKey)

	isRequired := false
	mappings := []model.EntitySchemasParametersMappings{
		{ParameterKey: pKey, IsRequired: &isRequired},
	}
	_, err := storage.CreateSchema(context.Background(), newDraftSchema(sKey, "Dup"), mappings)
	require.NoError(t, err, "first create should succeed")

	_, err = storage.CreateSchema(context.Background(), newDraftSchema(sKey, "Dup"), mappings)
	require.Error(t, err, "second create should violate the single-draft unique index")
}

// --- DeleteSchemaVersion ---

func TestDeleteSchemaVersion_Draft(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	pKey := uniqueKey("p")
	sKey := uniqueKey("s")
	seedParameter(t, db, pKey)
	cleanupSchema(t, db, sKey)

	isRequired := false
	_, err := storage.CreateSchema(context.Background(),
		newDraftSchema(sKey, "Del Schema"),
		[]model.EntitySchemasParametersMappings{{ParameterKey: pKey, IsRequired: &isRequired}},
	)
	require.NoError(t, err)

	require.NoError(t, storage.DeleteSchemaVersion(context.Background(), sKey, DraftVersion))

	_, err = storage.GetSchemaVersion(context.Background(), sKey, DraftVersion)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema version not found")
}

func TestDeleteSchemaVersion_ActiveIsRejected(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	pKey := uniqueKey("p")
	sKey := uniqueKey("s")
	seedParameter(t, db, pKey)
	cleanupSchema(t, db, sKey)

	isRequired := false
	_, err := storage.CreateSchema(context.Background(),
		newDraftSchema(sKey, "Active Schema"),
		[]model.EntitySchemasParametersMappings{{ParameterKey: pKey, IsRequired: &isRequired}},
	)
	require.NoError(t, err)

	_, err = storage.PublishDraft(context.Background(), sKey)
	require.NoError(t, err)

	err = storage.DeleteSchemaVersion(context.Background(), sKey, "1.0.0")
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "schema_not_draft:"),
		"expected schema_not_draft: prefix, got: %s", err.Error())
}

// --- PublishDraft ---

func TestPublishDraft_FirstPublish(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	pKey := uniqueKey("p")
	sKey := uniqueKey("s")
	seedParameter(t, db, pKey)
	cleanupSchema(t, db, sKey)

	isRequired := true
	_, err := storage.CreateSchema(context.Background(),
		newDraftSchema(sKey, "First Pub"),
		[]model.EntitySchemasParametersMappings{{ParameterKey: pKey, IsRequired: &isRequired}},
	)
	require.NoError(t, err)

	result, err := storage.PublishDraft(context.Background(), sKey)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "1.0.0", result.SchemaVersion)
	assert.Equal(t, "major", result.VersionBump)
	assert.Empty(t, result.PreviousVersion)
	assert.Equal(t, "active", *result.Lifecycle)
	assert.True(t, *result.IsLatest)
	require.Len(t, result.Parameters, 1)
	assert.Equal(t, pKey, result.Parameters[0].ParameterKey)
}

func TestPublishDraft_VersionBumpDetection(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	pOpt := uniqueKey("o") // optional param
	pReq := uniqueKey("r") // required param (added in second draft)
	sKey := uniqueKey("s")
	seedParameter(t, db, pOpt)
	seedParameter(t, db, pReq)
	cleanupSchema(t, db, sKey)

	// First publish: schema with one optional parameter → 1.0.0
	isOpt := false
	_, err := storage.CreateSchema(context.Background(),
		newDraftSchema(sKey, "Bump Test"),
		[]model.EntitySchemasParametersMappings{{ParameterKey: pOpt, IsRequired: &isOpt}},
	)
	require.NoError(t, err)
	first, err := storage.PublishDraft(context.Background(), sKey)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", first.SchemaVersion)

	// Second draft: add a new required parameter → should trigger major bump → 2.0.0
	isReq := true
	_, err = storage.CreateSchema(context.Background(),
		newDraftSchema(sKey, "Bump Test"),
		[]model.EntitySchemasParametersMappings{
			{ParameterKey: pOpt, IsRequired: &isOpt},
			{ParameterKey: pReq, IsRequired: &isReq},
		},
	)
	require.NoError(t, err)
	second, err := storage.PublishDraft(context.Background(), sKey)
	require.NoError(t, err)

	assert.Equal(t, "2.0.0", second.SchemaVersion)
	assert.Equal(t, "major", second.VersionBump)
	assert.Equal(t, "1.0.0", second.PreviousVersion)
	assert.True(t, *second.IsLatest)
}

func TestPublishDraft_NoDraft(t *testing.T) {
	db := requireDB(t)
	storage := NewStorage(db)

	_, err := storage.PublishDraft(context.Background(), "no_such_schema_key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no draft found")
}
