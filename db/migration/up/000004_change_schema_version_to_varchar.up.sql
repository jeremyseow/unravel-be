ALTER TABLE entity_schemas ALTER COLUMN schema_version TYPE VARCHAR(32) USING schema_version::text;
ALTER TABLE entity_schemas_parameters_mappings ALTER COLUMN schema_version TYPE VARCHAR(32) USING schema_version::text;
