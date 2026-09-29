// Package material provides the material domain types.
package material

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Type classifies a material.
type Type string

// Supported material classifications.
const (
	Unspecified     Type = "unspecified"
	RawMaterial     Type = "raw_material"
	Packaging       Type = "packaging"
	SemiFinished    Type = "semi_finished"
	FinishedProduct Type = "finished_product"
	SparePart       Type = "spare_part"
)

// Valid reports whether t is a supported classification.
func (t Type) Valid() bool {
	switch t {
	case Unspecified, RawMaterial, Packaging, SemiFinished, FinishedProduct, SparePart:
		return true
	}

	return false
}

// MaxNameLength matches the materials.name column.
const MaxNameLength = 255

// Material is a stock item. ID is database-generated.
type Material struct {
	ID   int64
	Name string
	Type Type
}

// Normalize trims the name.
func (m *Material) Normalize() {
	m.Name = strings.TrimSpace(m.Name)
}

// Validate normalizes then rejects a material with a blank or overlong name,
// or an unknown type.
func (m *Material) Validate() error {
	m.Normalize()

	if m.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalid)
	}

	if utf8.RuneCountInString(m.Name) > MaxNameLength {
		return fmt.Errorf("%w: name exceeds %d characters", ErrInvalid, MaxNameLength)
	}

	if !m.Type.Valid() {
		return fmt.Errorf("%w: unknown type %q", ErrInvalid, m.Type)
	}

	return nil
}
