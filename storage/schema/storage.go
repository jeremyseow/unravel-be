package schema

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

// validateParameterKeys checks that every key in the supplied list exists in
// entity_parameters for the default tenant. It runs inside the provided
// transaction so the check is consistent with the subsequent inserts.
func validateParameterKeys(tx *sql.Tx, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	exprs := make([]Expression, len(keys))
	for i, k := range keys {
		exprs[i] = String(k)
	}

	stmt := SELECT(EntityParameters.ParameterKey).
		FROM(EntityParameters).
		WHERE(
			EntityParameters.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntityParameters.ParameterKey.IN(exprs...)),
		)

	var found []model.EntityParameters
	if err := stmt.Query(tx, &found); err != nil {
		return err
	}

	foundSet := make(map[string]bool, len(found))
	for _, p := range found {
		foundSet[p.ParameterKey] = true
	}

	var missing []string
	for _, k := range keys {
		if !foundSet[k] {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("parameter keys not found in catalog: %v", missing)
	}
	return nil
}

// TODO: Replace with tenant ID from auth context (Phase 6)
var defaultTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// SchemaRecord holds a schema row together with its parameter mappings.
type SchemaRecord struct {
	model.EntitySchemas
	Parameters []model.EntitySchemasParametersMappings
}

type Storage interface {
	CreateSchema(ctx context.Context, schema model.EntitySchemas, mappings []model.EntitySchemasParametersMappings) (*SchemaRecord, error)
	GetSchemasByKey(ctx context.Context, key string) ([]SchemaRecord, error)
	GetSchemaVersion(ctx context.Context, key string, version string) (*SchemaRecord, error)
}

type StorageImpl struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) Storage {
	return &StorageImpl{db: db}
}

func (s *StorageImpl) CreateSchema(_ context.Context, schema model.EntitySchemas, mappings []model.EntitySchemasParametersMappings) (*SchemaRecord, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	schema.TenantID = defaultTenantID

	keys := make([]string, len(mappings))
	for i, m := range mappings {
		keys[i] = m.ParameterKey
	}
	if err = validateParameterKeys(tx, keys); err != nil {
		return nil, err
	}

	insertSchema := EntitySchemas.INSERT(
		EntitySchemas.TenantID,
		EntitySchemas.SchemaKey,
		EntitySchemas.SchemaName,
		EntitySchemas.SchemaVersion,
		EntitySchemas.Description,
		EntitySchemas.IsLatest,
		EntitySchemas.Lifecycle,
	).VALUES(
		schema.TenantID,
		schema.SchemaKey,
		schema.SchemaName,
		schema.SchemaVersion,
		schema.Description,
		schema.IsLatest,
		schema.Lifecycle,
	).RETURNING(EntitySchemas.AllColumns)

	var created model.EntitySchemas
	if err = insertSchema.Query(tx, &created); err != nil {
		return nil, err
	}

	var createdMappings []model.EntitySchemasParametersMappings
	for _, m := range mappings {
		m.TenantID = defaultTenantID
		m.SchemaKey = created.SchemaKey
		m.SchemaVersion = created.SchemaVersion

		insertMapping := EntitySchemasParametersMappings.INSERT(
			EntitySchemasParametersMappings.TenantID,
			EntitySchemasParametersMappings.SchemaKey,
			EntitySchemasParametersMappings.SchemaVersion,
			EntitySchemasParametersMappings.ParameterKey,
			EntitySchemasParametersMappings.IsRequired,
		).VALUES(
			m.TenantID,
			m.SchemaKey,
			m.SchemaVersion,
			m.ParameterKey,
			m.IsRequired,
		).RETURNING(EntitySchemasParametersMappings.AllColumns)

		var createdMapping model.EntitySchemasParametersMappings
		if err = insertMapping.Query(tx, &createdMapping); err != nil {
			return nil, err
		}
		createdMappings = append(createdMappings, createdMapping)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &SchemaRecord{EntitySchemas: created, Parameters: createdMappings}, nil
}

func (s *StorageImpl) GetSchemasByKey(_ context.Context, key string) ([]SchemaRecord, error) {
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntitySchemas.SchemaKey.EQ(String(key))),
		)

	var schemas []model.EntitySchemas
	if err := stmt.Query(s.db, &schemas); err != nil {
		return nil, err
	}
	if len(schemas) == 0 {
		return nil, fmt.Errorf("schema not found: %s", key)
	}

	records := make([]SchemaRecord, 0, len(schemas))
	for _, sc := range schemas {
		mappings, err := s.getMappings(sc.SchemaKey, sc.SchemaVersion)
		if err != nil {
			return nil, err
		}
		records = append(records, SchemaRecord{EntitySchemas: sc, Parameters: mappings})
	}

	sortByVersion(records)
	return records, nil
}

// sortByVersion sorts SchemaRecords newest-first using semver comparison.
// Handles major.minor.patch today; extend compareSemver for pre-release tags
// when the CHECK constraint is relaxed in a future migration.
func sortByVersion(records []SchemaRecord) {
	sort.Slice(records, func(i, j int) bool {
		return compareSemver(records[i].SchemaVersion, records[j].SchemaVersion) > 0
	})
}

// compareSemver returns 1 if a > b, -1 if a < b, 0 if equal.
// Supports optional pre-release suffix (e.g. "1.0.0-rc.1"):
//   - release (no suffix) ranks above any pre-release
//   - pre-release identifiers are compared lexicographically
func compareSemver(a, b string) int {
	parseVersion := func(v string) (major, minor, patch int, pre string) {
		parts := strings.SplitN(v, "-", 2)
		nums := strings.Split(parts[0], ".")
		toInt := func(s string) int { n, _ := strconv.Atoi(s); return n }
		if len(nums) > 0 {
			major = toInt(nums[0])
		}
		if len(nums) > 1 {
			minor = toInt(nums[1])
		}
		if len(nums) > 2 {
			patch = toInt(nums[2])
		}
		if len(parts) == 2 {
			pre = parts[1]
		}
		return
	}

	aMaj, aMin, aPat, aPre := parseVersion(a)
	bMaj, bMin, bPat, bPre := parseVersion(b)

	for _, pair := range [3][2]int{{aMaj, bMaj}, {aMin, bMin}, {aPat, bPat}} {
		if pair[0] != pair[1] {
			if pair[0] > pair[1] {
				return 1
			}
			return -1
		}
	}

	// Release (empty pre) ranks above any pre-release
	switch {
	case aPre == "" && bPre != "":
		return 1
	case aPre != "" && bPre == "":
		return -1
	case aPre != bPre:
		if aPre > bPre {
			return 1
		}
		return -1
	}
	return 0
}

func (s *StorageImpl) GetSchemaVersion(_ context.Context, key string, version string) (*SchemaRecord, error) {
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntitySchemas.SchemaKey.EQ(String(key))).
				AND(EntitySchemas.SchemaVersion.EQ(String(version))),
		).
		LIMIT(1)

	var schemas []model.EntitySchemas
	if err := stmt.Query(s.db, &schemas); err != nil {
		return nil, err
	}
	if len(schemas) == 0 {
		return nil, fmt.Errorf("schema version not found: %s@%s", key, version)
	}

	mappings, err := s.getMappings(schemas[0].SchemaKey, schemas[0].SchemaVersion)
	if err != nil {
		return nil, err
	}
	return &SchemaRecord{EntitySchemas: schemas[0], Parameters: mappings}, nil
}

func (s *StorageImpl) getMappings(schemaKey, schemaVersion string) ([]model.EntitySchemasParametersMappings, error) {
	stmt := SELECT(EntitySchemasParametersMappings.AllColumns).
		FROM(EntitySchemasParametersMappings).
		WHERE(
			EntitySchemasParametersMappings.TenantID.EQ(String(defaultTenantID.String())).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(schemaKey))).
				AND(EntitySchemasParametersMappings.SchemaVersion.EQ(String(schemaVersion))),
		)

	var mappings []model.EntitySchemasParametersMappings
	err := stmt.Query(s.db, &mappings)
	return mappings, err
}
