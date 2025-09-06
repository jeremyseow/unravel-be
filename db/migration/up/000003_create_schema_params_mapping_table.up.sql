CREATE TABLE entity_schemas_parameters_mappings (
	id BIGSERIAL,
	tenant_id uuid NOT NULL,
	schema_key VARCHAR(32) NOT NULL,
	schema_version jsonb NOT NULL,
	parameter_key VARCHAR(32) NOT NULL,
	is_required BOOLEAN DEFAULT TRUE,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	
	PRIMARY KEY (id)
);

CREATE UNIQUE INDEX entity_mappings_tenant_schema_parameter_uni_idx ON entity_schemas_parameters_mappings (tenant_id, schema_key, schema_version, parameter_key);