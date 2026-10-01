// Package material provides the material domain types.
package material

import (
	"erp/internal/db"
	"fmt"
	"slices"
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
	Consumable      Type = "consumable"
)

// Valid reports whether t is a supported classification.
func (t Type) Valid() bool {
	switch t {
	case Unspecified,
		RawMaterial,
		Packaging,
		SemiFinished,
		FinishedProduct,
		SparePart,
		Consumable:
		return true
	}

	return false
}

// Profile lists the characteristic names the type allows.
// Unspecified allows none: specs need a type first.
func (t Type) Profile() []SpecName {
	switch t {
	case RawMaterial:
		return []SpecName{
			SpecGrade,
			SpecMFI,
			SpecDensity,
			SpecAdditive,
			SpecConcentration,
		}
	case SemiFinished:
		return []SpecName{
			SpecThickness,
			SpecWidth,
			SpecLength,
			SpecWeight,
		}
	case FinishedProduct:
		return []SpecName{
			SpecGrade,
			SpecThickness,
			SpecWidth,
			SpecLength,
			SpecWeight,
			SpecColor,
			SpecTensile,
			SpecElongation,
			SpecShrinkage,
			SpecCOF,
			SpecHaze,
			SpecGloss,
			SpecTreatment,
			SpecSealStrength,
			SpecSealTemp,
			SpecWVTR,
			SpecOTR,
		}
	case Packaging:
		return []SpecName{
			SpecLength,
			SpecWidth,
			SpecHeight,
			SpecFlute,
			SpecStrength,
			SpecInnerDiameter,
			SpecWallThickness,
		}
	case SparePart:
		return []SpecName{
			SpecPartNumber,
			SpecDimensions,
			SpecOEM,
		}
	case Consumable:
		return []SpecName{
			SpecGrade,
			SpecApplication,
		}
	default:
		return nil
	}
}

// MaxSpecValueLength matches the material_specs.value column.
const MaxSpecValueLength = 255

// SpecName is a predefined characteristic name.
type SpecName string

// Supported characteristic names.
const (
	SpecGrade         SpecName = "grade"
	SpecMFI           SpecName = "mfi"
	SpecDensity       SpecName = "density"
	SpecAdditive      SpecName = "additive"
	SpecConcentration SpecName = "concentration"
	SpecThickness     SpecName = "thickness"
	SpecWidth         SpecName = "width"
	SpecLength        SpecName = "length"
	SpecWeight        SpecName = "weight"
	SpecColor         SpecName = "color"
	SpecTensile       SpecName = "tensile"
	SpecElongation    SpecName = "elongation"
	SpecShrinkage     SpecName = "shrinkage"
	SpecCOF           SpecName = "cof"
	SpecHaze          SpecName = "haze"
	SpecGloss         SpecName = "gloss"
	SpecTreatment     SpecName = "treatment"
	SpecSealStrength  SpecName = "seal_strength"
	SpecSealTemp      SpecName = "seal_temp"
	SpecWVTR          SpecName = "wvtr"
	SpecOTR           SpecName = "otr"
	SpecHeight        SpecName = "height"
	SpecFlute         SpecName = "flute"
	SpecStrength      SpecName = "strength"
	SpecInnerDiameter SpecName = "inner_diameter"
	SpecWallThickness SpecName = "wall_thickness"
	SpecPartNumber    SpecName = "part_number"
	SpecDimensions    SpecName = "dimensions"
	SpecOEM           SpecName = "oem"
	SpecApplication   SpecName = "application"
)

// Spec is one recorded name → value pair on a material.
type Spec struct {
	Name  SpecName
	Value string
}

// Normalize trims the name and value.
func (s *Spec) Normalize() {
	s.Name = SpecName(strings.TrimSpace(string(s.Name)))
	s.Value = strings.TrimSpace(s.Value)
}

// Validate normalizes then rejects a blank or overlong value,
// or a name outside the allowed profile.
func (s *Spec) Validate(allowed []SpecName) error {
	s.Normalize()

	if s.Name == "" {
		return fmt.Errorf("%w: spec name is required", ErrInvalid)
	}

	if s.Value == "" {
		return fmt.Errorf("%w: spec value is required", ErrInvalid)
	}

	if strings.ContainsRune(s.Value, 0) {
		return fmt.Errorf("%w: spec value contains a NUL byte", ErrInvalid)
	}

	if utf8.RuneCountInString(s.Value) > MaxSpecValueLength {
		return fmt.Errorf(
			"%w: spec value exceeds %d characters",
			ErrInvalid,
			MaxSpecValueLength,
		)
	}

	if !slices.Contains(allowed, s.Name) {
		return fmt.Errorf("%w: spec %q outside the profile", ErrInvalid, s.Name)
	}

	return nil
}

// Specs is the recorded set for one material.
type Specs []Spec

// Merge returns the combined set without modifying either operand.
func (s Specs) Merge(other Specs) Specs {
	merged := make(Specs, 0, len(s)+len(other))
	merged = append(merged, s...)
	merged = append(merged, other...)

	return merged
}

// Columns splits the set into parallel name and value slices
// for bulk inserts.
func (s Specs) Columns() ([]string, []string) {
	names := make([]string, 0, len(s))
	values := make([]string, 0, len(s))

	for _, spec := range s {
		names = append(names, string(spec.Name))
		values = append(values, spec.Value)
	}

	return names, values
}

// Validate checks every spec against the type profile and rejects
// case-insensitive duplicates.
func (s Specs) Validate(t Type) error {
	allowed := t.Profile()
	seen := make([]string, 0, len(s))

	for i := range s {
		if err := s[i].Validate(allowed); err != nil {
			return err
		}

		lowered := strings.ToLower(string(s[i].Name))
		if slices.Contains(seen, lowered) {
			return fmt.Errorf("%w: duplicated spec %q", ErrInvalid, s[i].Name)
		}

		seen = append(seen, lowered)
	}

	return nil
}

// MaxNameLength matches the materials.name column.
const MaxNameLength = 255

// Material is a stock item. ID is database-generated.
type Material struct {
	ID    int64
	Name  string
	Type  Type
	Specs Specs
}

// Normalize trims the name.
func (m *Material) Normalize() {
	m.Name = strings.TrimSpace(m.Name)
}

// AttachSpecs converts spec rows onto the material.
func (m *Material) AttachSpecs(rows []db.GetMaterialSpecsRow) {
	for _, row := range rows {
		m.Specs = append(m.Specs, Spec{
			Name:  SpecName(row.Name),
			Value: row.Value,
		})
	}
}

// Validate normalizes then rejects a blank or overlong name,
// an unknown type, or invalid specs.
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

	return m.Specs.Validate(m.Type)
}
