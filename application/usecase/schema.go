package usecase

import (
	"context"
	"fmt"

	"github.com/jeremyseow/unravel-be/application/domain"
)

// DraftVersion is the semver placeholder for every new draft.
// The DB check constraint requires semver format, so 0.0.0 is used.
const DraftVersion = "0.0.0"

type SchemaService interface {
	CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error)
	GetSchemas(ctx context.Context, key string) ([]domain.Schema, error)
	GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error)
}

type SchemaRepository interface {
	CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error)
	GetSchemas(ctx context.Context, key string) ([]domain.Schema, error)
	GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error)
	GetParametersByKeys(ctx context.Context, keys []string) ([]domain.Parameter, error)
}

type schemaService struct {
	repo SchemaRepository
}

func NewSchemaService(repo SchemaRepository) SchemaService {
	return &schemaService{repo: repo}
}

func (s *schemaService) CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error) {
	// Validate that all referenced parameter keys exist in the catalog.
	if len(schema.Parameters) > 0 {
		keys := make([]string, len(schema.Parameters))
		for i, p := range schema.Parameters {
			keys[i] = p.ParameterKey
		}
		found, err := s.repo.GetParametersByKeys(ctx, keys)
		if err != nil {
			return domain.Schema{}, fmt.Errorf("validating parameters: %w", err)
		}
		if len(found) != len(keys) {
			return domain.Schema{}, fmt.Errorf("parameter keys not found in catalog")
		}
	}

	schema.SchemaVersion = DraftVersion

	return s.repo.CreateSchema(ctx, schema)
}

func (s *schemaService) GetSchemas(ctx context.Context, key string) ([]domain.Schema, error) {
	return s.repo.GetSchemas(ctx, key)
}

func (s *schemaService) GetSchemaVersion(ctx context.Context, key, version string) (domain.Schema, error) {
	return s.repo.GetSchemaVersion(ctx, key, version)
}
