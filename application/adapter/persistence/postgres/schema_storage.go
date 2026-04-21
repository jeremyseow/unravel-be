package postgres

import (
	"context"
	"database/sql"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jeremyseow/unravel-be/application/domain"
	"github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/model"
	. "github.com/jeremyseow/unravel-be/db/.gen/unravel-db/public/table"
)

type SchemaStorage struct {
	db *sql.DB
}

func NewSchemaStorage(db *sql.DB) *SchemaStorage {
	return &SchemaStorage{
		db: db,
	}
}

func (s *SchemaStorage) CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error) {
	// For now, we will use a nil tenant_id
	tenantID := uuid.Nil

	// Insert into entity_schemas
	stmt := EntitySchemas.INSERT(
		EntitySchemas.TenantID,
		EntitySchemas.SchemaKey,
		EntitySchemas.SchemaVersion,
		EntitySchemas.Description,
	).VALUES(
		tenantID,
		schema.SchemaKey,
		schema.SchemaVersion,
		schema.Description,
	).RETURNING(EntitySchemas.AllColumns)

	var dbSchema model.EntitySchemas
	err := stmt.QueryContext(ctx, s.db, &dbSchema)
	if err != nil {
		return domain.Schema{}, err
	}

	// Insert into entity_schemas_parameters_mappings
	if len(schema.Parameters) > 0 {
		insertStmt := EntitySchemasParametersMappings.INSERT(
			EntitySchemasParametersMappings.TenantID,
			EntitySchemasParametersMappings.SchemaKey,
			EntitySchemasParametersMappings.SchemaVersion,
			EntitySchemasParametersMappings.ParameterKey,
		)
		for _, paramKey := range schema.Parameters {
			insertStmt = insertStmt.VALUES(
				tenantID,
				schema.SchemaKey,
				schema.SchemaVersion,
				paramKey,
			)
		}
		_, err = insertStmt.ExecContext(ctx, s.db)
		if err != nil {
			// TODO: Handle potential rollback
			return domain.Schema{}, err
		}
	}

	return toDomainSchema(dbSchema, schema.Parameters), nil
}

func (s *SchemaStorage) GetSchemas(ctx context.Context, name string) ([]domain.Schema, error) {
	// For now, we will use a nil tenant_id
	tenantID := uuid.Nil

	stmt := SELECT(
		EntitySchemas.AllColumns,
	).FROM(
		EntitySchemas,
	).WHERE(
		EntitySchemas.TenantID.EQ(UUID(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(name))),
	)

	var dbSchemas []model.EntitySchemas
	err := stmt.QueryContext(ctx, s.db, &dbSchemas)
	if err != nil {
		return nil, err
	}

	var domainSchemas []domain.Schema
	for _, dbSchema := range dbSchemas {
		// This is inefficient, we should do a join. For now, we will do a separate query for each schema.
		params, err := s.getSchemaParameters(ctx, tenantID, dbSchema.SchemaKey, dbSchema.SchemaVersion)
		if err != nil {
			return nil, err
		}
		domainSchemas = append(domainSchemas, toDomainSchema(dbSchema, params))
	}

	return domainSchemas, nil
}

func (s *SchemaStorage) GetSchemaVersion(ctx context.Context, name, version string) (domain.Schema, error) {
	// For now, we will use a nil tenant_id
	tenantID := uuid.Nil

	stmt := SELECT(
		EntitySchemas.AllColumns,
	).FROM(
		EntitySchemas,
	).WHERE(
		EntitySchemas.TenantID.EQ(UUID(tenantID)).
			AND(EntitySchemas.SchemaKey.EQ(String(name))).
			AND(EntitySchemas.SchemaVersion.EQ(String(version))),
	)

	var dbSchema model.EntitySchemas
	err := stmt.QueryContext(ctx, s.db, &dbSchema)
	if err != nil {
		return domain.Schema{}, err
	}

	params, err := s.getSchemaParameters(ctx, tenantID, dbSchema.SchemaKey, dbSchema.SchemaVersion)
	if err != nil {
		return domain.Schema{}, err
	}

	return toDomainSchema(dbSchema, params), nil
}

func (s *SchemaStorage) getSchemaParameters(ctx context.Context, tenantID uuid.UUID, schemaKey, schemaVersion string) ([]string, error) {
	stmt := SELECT(
		EntitySchemasParametersMappings.ParameterKey,
	).FROM(
		EntitySchemasParametersMappings,
	).WHERE(
		EntitySchemasParametersMappings.TenantID.EQ(UUID(tenantID)).
			AND(EntitySchemasParametersMappings.SchemaKey.EQ(String(schemaKey))).
			AND(EntitySchemasParametersMappings.SchemaVersion.EQ(String(schemaVersion))),
	)

	var paramKeys []string
	err := stmt.QueryContext(ctx, s.db, &paramKeys)
	if err != nil {
		return nil, err
	}

	return paramKeys, nil
}

func toDomainSchema(dbSchema model.EntitySchemas, params []string) domain.Schema {
	return domain.Schema{
		ID:            dbSchema.ID,
		TenantID:      dbSchema.TenantID,
		SchemaKey:     dbSchema.SchemaKey,
		SchemaVersion: dbSchema.SchemaVersion,
		Description:   dbSchema.Description,
		Parameters:    params,
		CreatedAt:     dbSchema.CreatedAt,
		UpdatedAt:     dbSchema.UpdatedAt,
	}
}
