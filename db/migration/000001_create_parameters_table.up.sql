create table entity_parameters (
	id BIGSERIAL,
	tenant_id uuid NOT NULL,
	parameter_key VARCHAR(32) NOT NULL,
	data_type VARCHAR(16) NOT NULL,
	description VARCHAR(255),
	created_at TIMESTAMP DEFAULT current_timestamp,
	updated_at TIMESTAMP DEFAULT current_timestamp,
	
	PRIMARY KEY (id)
);

CREATE UNIQUE INDEX entity_parameters_tenant_parameter_uni_idx ON entity_parameters (tenant_id, parameter_key);

CREATE INDEX entity_parameters_data_type_idx ON entity_parameters (data_type);