package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jeremyseow/unravel-be/application/domain"
)

type SchemaService interface {
	CreateSchema(ctx context.Context, schema domain.Schema, params map[string]any) (domain.Schema, error)
	GetSchemas(ctx context.Context, name string) ([]domain.Schema, error)
	GetSchemaVersion(ctx context.Context, name, version string) (domain.Schema, error)
}

type SchemaRepository interface {
	CreateSchema(ctx context.Context, schema domain.Schema) (domain.Schema, error)
	GetSchemas(ctx context.Context, name string) ([]domain.Schema, error)
	GetSchemaVersion(ctx context.Context, name, version string) (domain.Schema, error)
}

type schemaService struct {
	repo SchemaRepository
}

func NewSchemaService(repo SchemaRepository) SchemaService {
	return &schemaService{
		repo: repo,
	}
}

func (s *schemaService) CreateSchema(ctx context.Context, schema domain.Schema, params map[string]any) (domain.Schema, error) {
	// Generate version based on parameters
	version := s.generateVersion(params)
	schema.SchemaVersion = version

	// Check if this version already exists
	_, err := s.repo.GetSchemaVersion(ctx, schema.SchemaKey, version)
	if err == nil {
		return domain.Schema{}, fmt.Errorf("schema version already exists")
	}

	// Prepare parameter keys
	var paramKeys []string
	for k := range params {
		paramKeys = append(paramKeys, k)
	}
	schema.Parameters = paramKeys

	// TODO: Get tenant ID from context if not provided
	if schema.TenantID == uuid.Nil {
		// Default or context-based tenant ID
	}

	return s.repo.CreateSchema(ctx, schema)
}

func (s *schemaService) GetSchemas(ctx context.Context, name string) ([]domain.Schema, error) {
	return s.repo.GetSchemas(ctx, name)
}

func (s *schemaService) GetSchemaVersion(ctx context.Context, name, version string) (domain.Schema, error) {
	return s.repo.GetSchemaVersion(ctx, name, version)
}

func (s *schemaService) generateVersion(params map[string]any) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	paramBytes, _ := json.Marshal(params)
	hash := sha256.Sum256(paramBytes)
	shortHash := hex.EncodeToString(hash[:])[:8]

	return "v1-" + shortHash
}
