CREATE TABLE entity_parameters (
	id BIGSERIAL,
	tenant_id uuid NOT NULL,
	parameter_key VARCHAR(32) NOT NULL,
	parameter_name VARCHAR(32) NOT NULL,
	data_type VARCHAR(16) NOT NULL,
	description VARCHAR(255),
	sample_values VARCHAR(255),
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	
	PRIMARY KEY (id)
);

CREATE UNIQUE INDEX entity_parameters_tenant_parameter_uni_idx ON entity_parameters (tenant_id, parameter_key);

CREATE INDEX entity_parameters_data_type_idx ON entity_parameters (data_type);