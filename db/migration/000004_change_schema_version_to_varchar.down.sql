ALTER TABLE entity_schemas ALTER COLUMN schema_version TYPE jsonb USING schema_version::jsonb;
ALTER TABLE entity_schemas_parameters_mappings ALTER COLUMN schema_version TYPE jsonb USING schema_version::jsonb;
