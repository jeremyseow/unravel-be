package postgres

import (
	"context"
	"database/sql"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres"
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
		defaultTenantID,
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
				defaultTenantID,
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
		EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
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
		EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(key))).
			AND(EntitySchemas.SchemaVersion.EQ(String(version))),
	)

	var rows []schemaWithMappings
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil || len(rows) == 0 {
		return domain.Schema{}, fmt.Errorf("schema version not found")
	}
	return toDomainSchema(rows[0].EntitySchemas, toSchemaParameters(rows[0].Parameters)), nil
}

func (s *SchemaStorage) GetParametersByKeys(ctx context.Context, keys []string) ([]domain.Parameter, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	keyExprs := make([]Expression, len(keys))
	for i, k := range keys {
		keyExprs[i] = String(k)
	}
	stmt := SELECT(EntityParameters.AllColumns).
		FROM(EntityParameters).
		WHERE(
			EntityParameters.TenantID.EQ(uuidStr(defaultTenantID)).
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
