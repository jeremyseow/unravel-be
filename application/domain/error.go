package domain

import "errors"

// DraftVersion is the semver placeholder used for every unpublished schema.
const DraftVersion = "0.0.0"

var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidDataType       = errors.New("invalid data type")
	ErrParameterKeysNotFound = errors.New("parameter keys not found in catalog")
	ErrSchemaNotDraft        = errors.New("schema is not a draft")
	ErrSchemaNotActive       = errors.New("schema is not active")
)
