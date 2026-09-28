package material

import "errors"

// ErrNotFound indicates that the requested material does not exist.
var ErrNotFound = errors.New("material not found")

// ErrInvalid indicates that a material failed validation.
var ErrInvalid = errors.New("invalid material")
