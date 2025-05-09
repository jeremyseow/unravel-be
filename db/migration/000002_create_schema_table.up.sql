create table entity_schemas (
	id BIGSERIAL,
	tenant_id uuid NOT NULL,
	schema_key VARCHAR(32) NOT NULL,
	schema_version jsonb NOT NULL,
	description VARCHAR(255),
	created_at TIMESTAMP DEFAULT current_timestamp,
	updated_at TIMESTAMP DEFAULT current_timestamp,
	
	PRIMARY KEY (id)
);

CREATE UNIQUE INDEX entity_schemas_tenant_schema_uni_idx ON entity_schemas (tenant_id, schema_key, schema_version);
