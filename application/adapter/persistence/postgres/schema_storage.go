package postgres

import (
	"context"
	"database/sql"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/jeremyseow/unravel-be/application/ctxkey"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

type SchemaStorage struct {
	db *sql.DB
}

func NewSchemaStorage(db *sql.DB) *SchemaStorage {
	return &SchemaStorage{db: db}
}

func (s *SchemaStorage) CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error) {
	tenantID := ctxkey.TenantID(ctx)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Schema{}, err
	}
	defer tx.Rollback() //nolint:errcheck

	isLatest := false
	lifecycle := "draft"

	insertSchema := EntitySchemas.INSERT(
		EntitySchemas.TenantID,
		EntitySchemas.SchemaKey,
		EntitySchemas.SchemaName_,
		EntitySchemas.SchemaVersion,
		EntitySchemas.Description,
		EntitySchemas.IsLatest,
		EntitySchemas.Lifecycle,
	).VALUES(
		tenantID,
		schema.SchemaKey,
		schema.SchemaName,
		schema.SchemaVersion,
		schema.Description,
		isLatest,
		lifecycle,
	).RETURNING(EntitySchemas.AllColumns)

	var dbSchema model.EntitySchemas
	if err := insertSchema.QueryContext(ctx, tx, &dbSchema); err != nil {
		return domain.Schema{}, err
	}

	if len(schema.Parameters) > 0 {
		insertMappings := EntitySchemasParametersMappings.INSERT(
			EntitySchemasParametersMappings.TenantID,
			EntitySchemasParametersMappings.SchemaKey,
			EntitySchemasParametersMappings.SchemaVersion,
			EntitySchemasParametersMappings.ParameterKey,
			EntitySchemasParametersMappings.IsRequired,
		)
		for _, p := range schema.Parameters {
			insertMappings = insertMappings.VALUES(
				tenantID,
				schema.SchemaKey,
				schema.SchemaVersion,
				p.ParameterKey,
				p.IsRequired,
			)
		}
		if _, err := insertMappings.ExecContext(ctx, tx); err != nil {
			return domain.Schema{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Schema{}, err
	}

	return toDomainSchema(dbSchema, schema.Parameters), nil
}

func (s *SchemaStorage) GetSchemas(ctx context.Context, key string) ([]domain.Schema, error) {
	tenantID := ctxkey.TenantID(ctx)
	stmt := SELECT(
		EntitySchemas.AllColumns,
		EntitySchemasParametersMappings.AllColumns,
	).FROM(
		EntitySchemas.LEFT_JOIN(
			EntitySchemasParametersMappings,
			EntitySchemasParametersMappings.TenantID.EQ(EntitySchemas.TenantID).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(EntitySchemas.SchemaKey)).
				AND(EntitySchemasParametersMappings.SchemaVersion.EQ(EntitySchemas.SchemaVersion)),
		),
	).WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))),
	)

	var rows []schemaWithMappings
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, err
	}

	schemas := make([]domain.Schema, len(rows))
	for i, row := range rows {
		schemas[i] = toDomainSchema(row.EntitySchemas, toSchemaParameters(row.Parameters))
	}
	return schemas, nil
}

func (s *SchemaStorage) GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error) {
	tenantID := ctxkey.TenantID(ctx)
	stmt := SELECT(
		EntitySchemas.AllColumns,
		EntitySchemasParametersMappings.AllColumns,
	).FROM(
		EntitySchemas.LEFT_JOIN(
			EntitySchemasParametersMappings,
			EntitySchemasParametersMappings.TenantID.EQ(EntitySchemas.TenantID).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(EntitySchemas.SchemaKey)).
				AND(EntitySchemasParametersMappings.SchemaVersion.EQ(EntitySchemas.SchemaVersion)),
		),
	).WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.SchemaVersion.EQ(String(version))),
	)

	var rows []schemaWithMappings
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil || len(rows) == 0 {
		return domain.Schema{}, fmt.Errorf("schema %q version %q: %w", key, version, domain.ErrNotFound)
	}
	return toDomainSchema(rows[0].EntitySchemas, toSchemaParameters(rows[0].Parameters)), nil
}

func (s *SchemaStorage) GetLatestActiveSchema(ctx context.Context, key string) (*domain.Schema, error) {
	tenantID := ctxkey.TenantID(ctx)
	stmt := SELECT(
		EntitySchemas.AllColumns,
		EntitySchemasParametersMappings.AllColumns,
	).FROM(
		EntitySchemas.LEFT_JOIN(
			EntitySchemasParametersMappings,
			EntitySchemasParametersMappings.TenantID.EQ(EntitySchemas.TenantID).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(EntitySchemas.SchemaKey)).
				AND(EntitySchemasParametersMappings.SchemaVersion.EQ(EntitySchemas.SchemaVersion)),
		),
	).WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.IsLatest.EQ(Bool(true))).
			AND(EntitySchemas.Lifecycle.EQ(String("active"))),
	)

	var rows []schemaWithMappings
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	result := toDomainSchema(rows[0].EntitySchemas, toSchemaParameters(rows[0].Parameters))
	return &result, nil
}

func (s *SchemaStorage) PublishSchema(ctx context.Context, key, newVersion string) (domain.Schema, error) {
	tenantID := ctxkey.TenantID(ctx)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Schema{}, err
	}
	defer tx.Rollback() //nolint:errcheck

	// Promote draft to active with the computed version.
	lifecycle := "active"
	isLatest := true
	updateSchema := EntitySchemas.UPDATE(
		EntitySchemas.SchemaVersion,
		EntitySchemas.Lifecycle,
		EntitySchemas.IsLatest,
	).SET(
		newVersion,
		lifecycle,
		isLatest,
	).WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.SchemaVersion.EQ(String(domain.DraftVersion))).
			AND(EntitySchemas.Lifecycle.EQ(String("draft"))),
	).RETURNING(EntitySchemas.AllColumns)

	var dbSchema model.EntitySchemas
	if err := updateSchema.QueryContext(ctx, tx, &dbSchema); err != nil {
		return domain.Schema{}, err
	}
	if dbSchema.SchemaKey == "" {
		return domain.Schema{}, fmt.Errorf("schema %q: %w", key, domain.ErrSchemaNotDraft)
	}

	// Update the version on all parameter mappings for this draft.
	if _, err := EntitySchemasParametersMappings.UPDATE(
		EntitySchemasParametersMappings.SchemaVersion,
	).SET(newVersion).WHERE(
		EntitySchemasParametersMappings.TenantID.EQ(uuidStr(tenantID)).
			AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(key))).
			AND(EntitySchemasParametersMappings.SchemaVersion.EQ(String(domain.DraftVersion))),
	).ExecContext(ctx, tx); err != nil {
		return domain.Schema{}, err
	}

	// Unset is_latest on the previous latest version.
	if _, err := EntitySchemas.UPDATE(EntitySchemas.IsLatest).SET(false).WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.IsLatest.EQ(Bool(true))).
			AND(EntitySchemas.SchemaVersion.NOT_EQ(String(newVersion))),
	).ExecContext(ctx, tx); err != nil {
		return domain.Schema{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Schema{}, err
	}

	// Re-fetch to include parameters.
	return s.GetSchemaVersion(ctx, key, newVersion)
}

func (s *SchemaStorage) DeprecateSchema(ctx context.Context, key, version string) (domain.Schema, error) {
	tenantID := ctxkey.TenantID(ctx)

	lifecycle := "deprecated"
	stmt := EntitySchemas.UPDATE(
		EntitySchemas.Lifecycle,
		EntitySchemas.IsLatest,
	).SET(
		lifecycle,
		false,
	).WHERE(
		EntitySchemas.TenantID.EQ(uuidStr(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.SchemaVersion.EQ(String(version))).
			AND(EntitySchemas.Lifecycle.EQ(String("active"))),
	).RETURNING(EntitySchemas.AllColumns)

	var dbSchema model.EntitySchemas
	if err := stmt.QueryContext(ctx, s.db, &dbSchema); err != nil {
		return domain.Schema{}, err
	}
	if dbSchema.SchemaKey == "" {
		return domain.Schema{}, fmt.Errorf("schema %q version %q: %w", key, version, domain.ErrSchemaNotActive)
	}

	return toDomainSchema(dbSchema, nil), nil
}

func (s *SchemaStorage) GetParametersByKeys(ctx context.Context, keys []string) ([]domain.Parameter, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	tenantID := ctxkey.TenantID(ctx)
	keyExprs := make([]Expression, len(keys))
	for i, k := range keys {
		keyExprs[i] = String(k)
	}
	stmt := SELECT(EntityParameters.AllColumns).
		FROM(EntityParameters).
		WHERE(
			EntityParameters.TenantID.EQ(uuidStr(tenantID)).
				AND(EntityParameters.ParameterKey.IN(keyExprs...)),
		)

	var rows []model.EntityParameters
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, err
	}

	params := make([]domain.Parameter, len(rows))
	for i, r := range rows {
		params[i] = toDomainParameter(r)
	}
	return params, nil
}

type schemaWithMappings struct {
	model.EntitySchemas
	Parameters []model.EntitySchemasParametersMappings
}

func toSchemaParameters(mappings []model.EntitySchemasParametersMappings) []domain.SchemaParameter {
	params := make([]domain.SchemaParameter, len(mappings))
	for i, m := range mappings {
		isRequired := m.IsRequired != nil && *m.IsRequired
		params[i] = domain.SchemaParameter{
			ParameterKey: m.ParameterKey,
			IsRequired:   isRequired,
		}
	}
	return params
}

func toDomainSchema(r model.EntitySchemas, params []domain.SchemaParameter) domain.Schema {
	return domain.Schema{
		ID:            r.ID,
		TenantID:      r.TenantID,
		SchemaKey:     r.SchemaKey,
		SchemaName:    r.SchemaName,
		SchemaVersion: r.SchemaVersion,
		Description:   r.Description,
		IsLatest:      r.IsLatest,
		Lifecycle:     r.Lifecycle,
		Parameters:    params,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
