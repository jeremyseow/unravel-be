package schema

import (
	"context"
	"database/sql"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

// DraftVersion is the placeholder semver stored for every schema draft.
// The real version (1.0.0, 2.1.0, …) is assigned at publish time.
// 0.0.0 is intentionally below any real version and can never be published.
const DraftVersion = "0.0.0"

// TODO: Replace with tenant ID from auth context (Phase 6)
var defaultTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// uuidStr casts a uuid.UUID to a StringExpression with an explicit ::uuid cast.
// Required because pq sends string parameters as the text OID, and PostgreSQL
// has no implicit uuid = text operator.
func uuidStr(id uuid.UUID) StringExpression {
	return StringExp(CAST(String(id.String())).AS("uuid"))
}

// SchemaRecord holds a schema row together with its parameter mappings.
type SchemaRecord struct {
	model.EntitySchemas
	Parameters []model.EntitySchemasParametersMappings
}

// PublishResult wraps the newly published SchemaRecord with version bump metadata.
type PublishResult struct {
	SchemaRecord
	// VersionBump is "major", "minor", or "patch".
	VersionBump string
	// PreviousVersion is the version that was superseded. Empty on first publish.
	PreviousVersion string
}

type Storage interface {
	CreateSchema(ctx context.Context, schema model.EntitySchemas, mappings []model.EntitySchemasParametersMappings) (*SchemaRecord, error)
	GetSchemasByKey(ctx context.Context, key string) ([]SchemaRecord, error)
	GetSchemaVersion(ctx context.Context, key string, version string) (*SchemaRecord, error)
	DeleteSchemaVersion(ctx context.Context, key string, version string) error
	PublishDraft(ctx context.Context, key string) (*PublishResult, error)
	DeprecateSchemaVersion(ctx context.Context, key string, version string) (*SchemaRecord, error)
}

type StorageImpl struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) Storage {
	return &StorageImpl{db: db}
}

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
			EntityParameters.TenantID.EQ(uuidStr(defaultTenantID)).
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
		EntitySchemas.SchemaName_,
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
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
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

func (s *StorageImpl) GetSchemaVersion(_ context.Context, key string, version string) (*SchemaRecord, error) {
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
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

func (s *StorageImpl) DeleteSchemaVersion(ctx context.Context, key, version string) error {
	record, err := s.GetSchemaVersion(ctx, key, version)
	if err != nil {
		return err
	}

	if record.Lifecycle != nil && *record.Lifecycle != "draft" {
		return fmt.Errorf("schema_not_draft: schema %s@%s has lifecycle %q; only draft schemas can be deleted", key, version, *record.Lifecycle)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	delMappings := EntitySchemasParametersMappings.DELETE().WHERE(
		EntitySchemasParametersMappings.TenantID.EQ(uuidStr(defaultTenantID)).
			AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(key))).
			AND(EntitySchemasParametersMappings.SchemaVersion.EQ(String(version))),
	)
	if _, err = delMappings.Exec(tx); err != nil {
		return err
	}

	delSchema := EntitySchemas.DELETE().WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.SchemaVersion.EQ(String(version))),
	)
	if _, err = delSchema.Exec(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// PublishDraft promotes the current draft to active, auto-detects the semver
// bump relative to the previous active version, and assigns the new version.
// First-time publishes always produce version 1.0.0 with bump "major".
func (s *StorageImpl) PublishDraft(_ context.Context, key string) (*PublishResult, error) {
	draft, err := s.getDraft(key)
	if err != nil {
		return nil, err
	}

	latestActive, err := s.getLatestActive(key)
	if err != nil {
		return nil, err
	}

	var newVersion, bump, previousVersion string
	if latestActive == nil {
		// First publish — assign 1.0.0 and classify the bump as "major" so
		// callers know this is a brand-new schema contract.
		newVersion = "1.0.0"
		bump = "major"
	} else {
		currentMap := toParamMap(latestActive.Parameters)
		nextMap := toParamMap(draft.Parameters)
		bump = determineVersionBump(currentMap, nextMap)
		newVersion = applyBump(latestActive.SchemaVersion, bump)
		previousVersion = latestActive.SchemaVersion
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Clear is_latest on the previous active version before promoting the draft.
	if latestActive != nil {
		unsetLatest := EntitySchemas.UPDATE(EntitySchemas.IsLatest).
			SET(false).
			WHERE(
				EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
					AND(EntitySchemas.SchemaKey.EQ(String(key))).
					AND(EntitySchemas.SchemaVersion.EQ(String(latestActive.SchemaVersion))),
			)
		if _, err = unsetLatest.Exec(tx); err != nil {
			return nil, err
		}
	}

	// Promote the draft: assign the computed version, flip to active + is_latest.
	var updatedSchemas []model.EntitySchemas
	promoteDraft := EntitySchemas.UPDATE(
		EntitySchemas.SchemaVersion,
		EntitySchemas.Lifecycle,
		EntitySchemas.IsLatest,
	).SET(
		newVersion,
		"active",
		true,
	).WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.SchemaVersion.EQ(String(DraftVersion))),
	).RETURNING(EntitySchemas.AllColumns)

	if err = promoteDraft.Query(tx, &updatedSchemas); err != nil {
		return nil, err
	}
	if len(updatedSchemas) == 0 {
		return nil, fmt.Errorf("no draft found for schema %s", key)
	}

	// Rewrite the mapping rows to point at the new version string.
	updateMappings := EntitySchemasParametersMappings.UPDATE(EntitySchemasParametersMappings.SchemaVersion).
		SET(newVersion).
		WHERE(
			EntitySchemasParametersMappings.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(key))).
				AND(EntitySchemasParametersMappings.SchemaVersion.EQ(String(DraftVersion))),
		)
	if _, err = updateMappings.Exec(tx); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	mappings, err := s.getMappings(updatedSchemas[0].SchemaKey, updatedSchemas[0].SchemaVersion)
	if err != nil {
		return nil, err
	}

	return &PublishResult{
		SchemaRecord:    SchemaRecord{EntitySchemas: updatedSchemas[0], Parameters: mappings},
		VersionBump:     bump,
		PreviousVersion: previousVersion,
	}, nil
}

// DeprecateSchemaVersion marks the given active version as deprecated.
// If it held the is_latest flag, the next-highest active version is promoted.
func (s *StorageImpl) DeprecateSchemaVersion(ctx context.Context, key, version string) (*SchemaRecord, error) {
	record, err := s.GetSchemaVersion(ctx, key, version)
	if err != nil {
		return nil, err
	}

	if record.Lifecycle == nil || *record.Lifecycle != "active" {
		lc := "unknown"
		if record.Lifecycle != nil {
			lc = *record.Lifecycle
		}
		return nil, fmt.Errorf("schema_not_active: schema %s@%s has lifecycle %q; only active schemas can be deprecated", key, version, lc)
	}

	wasLatest := record.IsLatest != nil && *record.IsLatest

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var updatedSchemas []model.EntitySchemas
	deprecateStmt := EntitySchemas.UPDATE(EntitySchemas.Lifecycle, EntitySchemas.IsLatest).
		SET("deprecated", false).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemas.SchemaKey.EQ(String(key))).
				AND(EntitySchemas.SchemaVersion.EQ(String(version))),
		).RETURNING(EntitySchemas.AllColumns)

	if err = deprecateStmt.Query(tx, &updatedSchemas); err != nil {
		return nil, err
	}

	// If this was the latest version, crown the next highest active version.
	if wasLatest {
		if err = s.promoteNextLatest(tx, key, version); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	if len(updatedSchemas) == 0 {
		return nil, fmt.Errorf("schema version not found: %s@%s", key, version)
	}

	mappings, err := s.getMappings(updatedSchemas[0].SchemaKey, updatedSchemas[0].SchemaVersion)
	if err != nil {
		return nil, err
	}

	return &SchemaRecord{EntitySchemas: updatedSchemas[0], Parameters: mappings}, nil
}

// getDraft returns the current draft for a schema key, or an error if none exists.
func (s *StorageImpl) getDraft(key string) (*SchemaRecord, error) {
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemas.SchemaKey.EQ(String(key))).
				AND(EntitySchemas.SchemaVersion.EQ(String(DraftVersion))),
		).LIMIT(1)

	var schemas []model.EntitySchemas
	if err := stmt.Query(s.db, &schemas); err != nil {
		return nil, err
	}
	if len(schemas) == 0 {
		return nil, fmt.Errorf("no draft found for schema %s", key)
	}

	mappings, err := s.getMappings(schemas[0].SchemaKey, schemas[0].SchemaVersion)
	if err != nil {
		return nil, err
	}
	return &SchemaRecord{EntitySchemas: schemas[0], Parameters: mappings}, nil
}

// getLatestActive returns the version currently flagged is_latest with lifecycle=active,
// or nil if no active version exists yet (first-time publish).
func (s *StorageImpl) getLatestActive(key string) (*SchemaRecord, error) {
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemas.SchemaKey.EQ(String(key))).
				AND(EntitySchemas.Lifecycle.EQ(String("active"))).
				AND(EntitySchemas.IsLatest.EQ(Bool(true))),
		).LIMIT(1)

	var schemas []model.EntitySchemas
	if err := stmt.Query(s.db, &schemas); err != nil {
		return nil, err
	}
	if len(schemas) == 0 {
		return nil, nil
	}

	mappings, err := s.getMappings(schemas[0].SchemaKey, schemas[0].SchemaVersion)
	if err != nil {
		return nil, err
	}
	return &SchemaRecord{EntitySchemas: schemas[0], Parameters: mappings}, nil
}

// promoteNextLatest finds the highest remaining active version for key
// (excluding excludeVersion) and sets its is_latest flag to true.
// Called inside a transaction after deprecating the previous latest version.
func (s *StorageImpl) promoteNextLatest(tx *sql.Tx, key, excludeVersion string) error {
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemas.SchemaKey.EQ(String(key))).
				AND(EntitySchemas.Lifecycle.EQ(String("active"))).
				AND(EntitySchemas.SchemaVersion.NOT_EQ(String(excludeVersion))),
		)

	var schemas []model.EntitySchemas
	if err := stmt.Query(tx, &schemas); err != nil {
		return err
	}
	if len(schemas) == 0 {
		return nil // no active version left to promote
	}

	// Sort in Go and pick the highest version.
	records := make([]SchemaRecord, len(schemas))
	for i, sc := range schemas {
		records[i] = SchemaRecord{EntitySchemas: sc}
	}
	sortByVersion(records)
	nextLatest := records[0].SchemaVersion

	updateStmt := EntitySchemas.UPDATE(EntitySchemas.IsLatest).
		SET(true).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemas.SchemaKey.EQ(String(key))).
				AND(EntitySchemas.SchemaVersion.EQ(String(nextLatest))),
		)
	_, err := updateStmt.Exec(tx)
	return err
}

func (s *StorageImpl) getMappings(schemaKey, schemaVersion string) ([]model.EntitySchemasParametersMappings, error) {
	stmt := SELECT(EntitySchemasParametersMappings.AllColumns).
		FROM(EntitySchemasParametersMappings).
		WHERE(
			EntitySchemasParametersMappings.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(schemaKey))).
				AND(EntitySchemasParametersMappings.SchemaVersion.EQ(String(schemaVersion))),
		)

	var mappings []model.EntitySchemasParametersMappings
	err := stmt.Query(s.db, &mappings)
	return mappings, err
}
