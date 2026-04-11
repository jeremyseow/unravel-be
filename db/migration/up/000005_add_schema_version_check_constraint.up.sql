ALTER TABLE entity_schemas
    ADD CONSTRAINT chk_schema_version_format
    CHECK (schema_version ~ '^\d+\.\d+\.\d+$');

ALTER TABLE entity_schemas_parameters_mappings
    ADD CONSTRAINT chk_mapping_version_format
    CHECK (schema_version ~ '^\d+\.\d+\.\d+$');
