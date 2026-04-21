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
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemas.SchemaKey.EQ(String(key))),
		)

	var rows []model.EntitySchemas
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, err
	}

	schemas := make([]domain.Schema, len(rows))
	for i, row := range rows {
		params, err := s.getSchemaParameters(ctx, row.SchemaKey, row.SchemaVersion)
		if err != nil {
			return nil, err
		}
		schemas[i] = toDomainSchema(row, params)
	}
	return schemas, nil
}

func (s *SchemaStorage) GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error) {
	stmt := SELECT(EntitySchemas.AllColumns).
		FROM(EntitySchemas).
		WHERE(
			EntitySchemas.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemas.SchemaKey.EQ(String(key))).
				AND(EntitySchemas.SchemaVersion.EQ(String(version))),
		)

	var row model.EntitySchemas
	if err := stmt.QueryContext(ctx, s.db, &row); err != nil {
		return domain.Schema{}, fmt.Errorf("schema version not found: %w", err)
	}

	params, err := s.getSchemaParameters(ctx, row.SchemaKey, row.SchemaVersion)
	if err != nil {
		return domain.Schema{}, err
	}
	return toDomainSchema(row, params), nil
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

func (s *SchemaStorage) getSchemaParameters(ctx context.Context, schemaKey, schemaVersion string) ([]domain.SchemaParameter, error) {
	stmt := SELECT(
		EntitySchemasParametersMappings.ParameterKey,
		EntitySchemasParametersMappings.IsRequired,
	).FROM(EntitySchemasParametersMappings).
		WHERE(
			EntitySchemasParametersMappings.TenantID.EQ(uuidStr(defaultTenantID)).
				AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(schemaKey))).
				AND(EntitySchemasParametersMappings.SchemaVersion.EQ(String(schemaVersion))),
		)

	var rows []model.EntitySchemasParametersMappings
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, err
	}

	params := make([]domain.SchemaParameter, len(rows))
	for i, r := range rows {
		isRequired := r.IsRequired != nil && *r.IsRequired
		params[i] = domain.SchemaParameter{
			ParameterKey: r.ParameterKey,
			IsRequired:   isRequired,
		}
	}
	return params, nil
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
