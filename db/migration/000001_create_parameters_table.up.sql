CREATE TABLE entity_parameters (
	id BIGSERIAL,
	tenant_id uuid NOT NULL,
	parameter_key VARCHAR(128) NOT NULL,
	parameter_name VARCHAR(128) NOT NULL,
	data_type VARCHAR(16) NOT NULL,
	description VARCHAR(255),
	sample_values TEXT,
	created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
	
	PRIMARY KEY (id),

	CONSTRAINT constraint_data_type_valid
    CHECK (data_type IN (
        'string', 'int32', 'int64', 'float', 'double', 'bool', 'bytes',
        'repeated_string', 'repeated_int32', 'repeated_int64',
        'repeated_float', 'repeated_double', 'repeated_bool'
    ))
);

CREATE UNIQUE INDEX entity_parameters_tenant_parameter_uni_idx ON entity_parameters (tenant_id, parameter_key);

CREATE INDEX entity_parameters_data_type_idx ON entity_parameters (data_type);