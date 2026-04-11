ALTER TABLE entity_parameters
    ADD CONSTRAINT chk_data_type_valid
    CHECK (data_type IN (
        'string', 'int32', 'int64', 'float', 'double', 'bool', 'bytes',
        'repeated_string', 'repeated_int32', 'repeated_int64',
        'repeated_float', 'repeated_double', 'repeated_bool'
    ));
