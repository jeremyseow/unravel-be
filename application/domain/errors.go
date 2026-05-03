package domain

import "errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrInvalidDataType       = errors.New("invalid data type")
	ErrParameterKeysNotFound = errors.New("parameter keys not found in catalog")
)
