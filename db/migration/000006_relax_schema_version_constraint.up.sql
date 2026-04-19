ALTER TABLE entity_schemas
    DROP CONSTRAINT chk_schema_version_format,
    ADD CONSTRAINT chk_schema_version_format
        CHECK (schema_version ~ '^\d+\.\d+\.\d+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$');

ALTER TABLE entity_schemas_parameters_mappings
    DROP CONSTRAINT chk_mapping_version_format,
    ADD CONSTRAINT chk_mapping_version_format
        CHECK (schema_version ~ '^\d+\.\d+\.\d+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$');
