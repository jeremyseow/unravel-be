CREATE TABLE entity_schemas (
	id BIGSERIAL,
	tenant_id uuid NOT NULL,
	schema_key VARCHAR(128) NOT NULL,
	schema_name VARCHAR(128) NOT NULL,
	schema_version VARCHAR(32) NOT NULL,
	description VARCHAR(255),
	is_latest BOOLEAN DEFAULT FALSE,
	lifecycle VARCHAR(16) DEFAULT 'draft',
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	
	PRIMARY KEY (id),

	CONSTRAINT constraint_schema_version_format
    CHECK (schema_version ~ '^\d+\.\d+\.\d+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$')
);

CREATE UNIQUE INDEX entity_schemas_tenant_schema_uni_idx ON entity_schemas (tenant_id, schema_key, schema_version);

CREATE UNIQUE INDEX entity_schemas_single_draft_idx ON entity_schemas (tenant_id, schema_key) WHERE lifecycle = 'draft';