ALTER TABLE entity_schemas
    DROP CONSTRAINT chk_schema_version_format;

ALTER TABLE entity_schemas_parameters_mappings
    DROP CONSTRAINT chk_mapping_version_format;
