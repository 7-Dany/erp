package material

import (
	"erp/internal/db"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func Test_Type_Valid(t *testing.T) {
	t.Run("accepts every defined type", func(t *testing.T) {
		defined := []Type{
			Unspecified, RawMaterial, Packaging,
			SemiFinished, FinishedProduct, SparePart, Consumable,
		}

		for _, typ := range defined {
			if !typ.Valid() {
				t.Errorf("Type(%q).Valid() = false, want true", typ)
			}
		}
	})

	t.Run("rejects the zero value", func(t *testing.T) {
		if Type("").Valid() {
			t.Error("Type(\"\").Valid() = true, want false")
		}
	})

	t.Run("rejects an unknown label", func(t *testing.T) {
		if Type("liquid_gold").Valid() {
			t.Error("unknown label reported as valid")
		}
	})
}

func Test_Type_CoversDatabaseEnum(t *testing.T) {
	t.Run("every database value is a valid domain type", func(t *testing.T) {
		for _, value := range db.AllMaterialTypeValues() {
			if !Type(value).Valid() {
				t.Errorf("Type(%q) from database has no domain constant", value)
			}
		}
	})

	t.Run("every domain type exists in the database enum", func(t *testing.T) {
		defined := []Type{
			Unspecified, RawMaterial, Packaging,
			SemiFinished, FinishedProduct, SparePart, Consumable,
		}

		for _, typ := range defined {
			if !db.MaterialType(typ).Valid() {
				t.Errorf("Type(%q) has no database enum value", typ)
			}
		}
	})
}

func Test_Type_Profile(t *testing.T) {
	t.Run("lists each type's characteristics", func(t *testing.T) {
		for typ, want := range map[Type][]SpecName{
			RawMaterial:     {SpecGrade, SpecMFI, SpecDensity, SpecAdditive, SpecConcentration},
			SemiFinished:    {SpecThickness, SpecWidth, SpecLength, SpecWeight},
			FinishedProduct: {SpecGrade, SpecThickness, SpecWidth, SpecLength, SpecWeight, SpecColor, SpecTensile, SpecElongation, SpecShrinkage, SpecCOF, SpecHaze, SpecGloss, SpecTreatment, SpecSealStrength, SpecSealTemp, SpecWVTR, SpecOTR},
			Packaging:       {SpecLength, SpecWidth, SpecHeight, SpecFlute, SpecStrength, SpecInnerDiameter, SpecWallThickness},
			SparePart:       {SpecPartNumber, SpecDimensions, SpecOEM},
			Consumable:      {SpecGrade, SpecApplication},
		} {
			got := typ.Profile()

			if len(got) != len(want) {
				t.Fatalf("Profile() type %q returned %d names, want %d", typ, len(got), len(want))
			}

			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("Profile() type %q = %v, want %v", typ, got, want)
				}
			}
		}
	})

	t.Run("returns nothing for unspecified and unknown types", func(t *testing.T) {
		if got := Unspecified.Profile(); len(got) != 0 {
			t.Errorf("Profile() unspecified = %v, want empty", got)
		}

		if got := Type("liquid_gold").Profile(); got != nil {
			t.Errorf("Profile() unknown = %v, want nil", got)
		}
	})
}

func Test_Material_AttachSpecs(t *testing.T) {
	t.Run("appends converted rows to the material", func(t *testing.T) {
		m := Material{Name: "BOPP 20UM", Type: FinishedProduct}

		m.AttachSpecs([]db.GetMaterialSpecsRow{{Name: "thickness", Value: "20 micron"}})

		want := Specs{{Name: "thickness", Value: "20 micron"}}
		if !reflect.DeepEqual(m.Specs, want) {
			t.Errorf("Specs = %+v, want %+v", m.Specs, want)
		}
	})

	t.Run("preserves already attached specs", func(t *testing.T) {
		m := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: Specs{
				{Name: "thickness", Value: "20 micron"},
			},
		}

		m.AttachSpecs([]db.GetMaterialSpecsRow{{Name: "tensile", Value: "130 MPa"}})

		if len(m.Specs) != 2 || m.Specs[0].Name != "thickness" || m.Specs[1].Name != "tensile" {
			t.Errorf("Specs = %+v, want both in order", m.Specs)
		}
	})
}

func Test_Material_Validate(t *testing.T) {
	t.Run("accepts a valid material", func(t *testing.T) {
		m := Material{Name: "Polypropylene Film", Type: RawMaterial}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("accepts a name at the maximum length", func(t *testing.T) {
		m := Material{Name: strings.Repeat("a", MaxNameLength), Type: RawMaterial}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("counts characters, not bytes, against the limit", func(t *testing.T) {
		m := Material{Name: strings.Repeat("é", MaxNameLength), Type: RawMaterial}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("rejects an empty name", func(t *testing.T) {
		m := Material{Name: "", Type: RawMaterial}
		err := m.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a whitespace-only name", func(t *testing.T) {
		m := Material{Name: "  \t ", Type: RawMaterial}
		err := m.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a name over the maximum length", func(t *testing.T) {
		name := strings.Repeat("a", MaxNameLength+1)

		m := Material{Name: name, Type: RawMaterial}
		err := m.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects an unset type", func(t *testing.T) {
		m := Material{Name: "Steel Sheet"}
		err := m.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects an unknown type", func(t *testing.T) {
		m := Material{Name: "Steel Sheet", Type: "liquid_gold"}
		err := m.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a whitespace-only type", func(t *testing.T) {
		m := Material{Name: "Steel Sheet", Type: "  "}
		err := m.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a mistyped-case type", func(t *testing.T) {
		m := Material{Name: "Steel Sheet", Type: "RAW_MATERIAL"}
		err := m.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("accepts unspecified as a type", func(t *testing.T) {
		m := Material{Name: "Mystery Drum", Type: Unspecified}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("accepts characteristics from the type profile", func(t *testing.T) {
		m := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: []Spec{
				{Name: "thickness", Value: "20 micron"},
				{Name: "tensile", Value: "130 MPa"},
			},
		}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("accepts a material without characteristics", func(t *testing.T) {
		m := Material{Name: "PP Resin", Type: RawMaterial}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("rejects a blank characteristic name or value", func(t *testing.T) {
		for _, specs := range [][]Spec{
			{{Name: "  ", Value: "20 micron"}},
			{{Name: "thickness", Value: "   "}},
		} {
			m := Material{Name: "BOPP 20UM", Type: FinishedProduct, Specs: specs}

			if err := m.Validate(); !errors.Is(err, ErrInvalid) {
				t.Errorf("Validate() specs %+v error = %v, want ErrInvalid", specs, err)
			}
		}
	})

	t.Run("rejects a duplicated characteristic name", func(t *testing.T) {
		m := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: []Spec{
				{Name: "color", Value: "clear"},
				{Name: "Color", Value: "matte"},
			},
		}

		if err := m.Validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a characteristic outside the type profile", func(t *testing.T) {
		for _, specs := range [][]Spec{
			{{Name: "tensile", Value: "130 MPa"}},
			{{Name: "TENSILE", Value: "130 MPa"}},
			{{Name: "color", Value: "clear"}},
		} {
			m := Material{Name: "PP Resin", Type: RawMaterial, Specs: specs}

			if err := m.Validate(); !errors.Is(err, ErrInvalid) {
				t.Errorf("Validate() specs %+v error = %v, want ErrInvalid", specs, err)
			}
		}
	})

	t.Run("trims characteristic names and values", func(t *testing.T) {
		m := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: []Spec{
				{Name: "  thickness  ", Value: "  20 micron "},
			},
		}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}

		if m.Specs[0].Name != "thickness" || m.Specs[0].Value != "20 micron" {
			t.Errorf("Validate() left specs %+v, want trimmed", m.Specs)
		}
	})

	t.Run("rejects a characteristic on unspecified", func(t *testing.T) {
		m := Material{
			Name: "Mystery Drum",
			Type: Unspecified,
			Specs: []Spec{
				{Name: "grade", Value: "HOMO-25"},
			},
		}

		if err := m.Validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("accepts a characteristic value at the maximum length", func(t *testing.T) {
		m := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: []Spec{
				{Name: "color", Value: strings.Repeat("a", MaxSpecValueLength)},
			},
		}

		if err := m.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("rejects a characteristic value with a NUL byte", func(t *testing.T) {
		m := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: []Spec{
				{Name: "color", Value: "cl\x00ear"},
			},
		}

		if err := m.Validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a characteristic value over the maximum length", func(t *testing.T) {
		m := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: []Spec{
				{Name: "color", Value: strings.Repeat("a", MaxSpecValueLength+1)},
			},
		}

		if err := m.Validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})
}

func Test_Spec_Normalize(t *testing.T) {
	t.Run("trims the name and value", func(t *testing.T) {
		s := Spec{Name: "  mfi ", Value: "  3 g/10min "}

		s.Normalize()

		if s.Name != "mfi" || s.Value != "3 g/10min" {
			t.Errorf("Normalize() left %+v, want trimmed", s)
		}
	})
}

func Test_Spec_Validate(t *testing.T) {
	allowed := RawMaterial.Profile()

	t.Run("accepts a profiled spec", func(t *testing.T) {
		s := Spec{Name: "mfi", Value: "3 g/10min"}

		if err := s.Validate(allowed); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("normalizes while validating", func(t *testing.T) {
		s := Spec{Name: "  mfi ", Value: "  3 g/10min "}

		if err := s.Validate(allowed); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}

		if s.Name != "mfi" || s.Value != "3 g/10min" {
			t.Errorf("Validate() left %+v, want trimmed", s)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		s := Spec{Name: "   ", Value: "3 g/10min"}

		if err := s.Validate(allowed); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a blank value", func(t *testing.T) {
		s := Spec{Name: "mfi", Value: "  "}

		if err := s.Validate(allowed); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a NUL byte in the value", func(t *testing.T) {
		s := Spec{Name: "mfi", Value: "3\x00"}

		if err := s.Validate(allowed); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a value over the maximum length", func(t *testing.T) {
		s := Spec{Name: "mfi", Value: strings.Repeat("a", MaxSpecValueLength+1)}

		if err := s.Validate(allowed); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a name outside the profile", func(t *testing.T) {
		s := Spec{Name: "tensile", Value: "130 MPa"}

		if err := s.Validate(allowed); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})
}

func Test_Specs_Validate(t *testing.T) {
	t.Run("accepts a valid set", func(t *testing.T) {
		s := Specs{
			{Name: "thickness", Value: "20 micron"},
			{Name: "tensile", Value: "130 MPa"},
		}

		if err := s.Validate(FinishedProduct); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("accepts an empty set", func(t *testing.T) {
		if err := Specs(nil).Validate(RawMaterial); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("rejects a case-insensitive duplicate", func(t *testing.T) {
		s := Specs{
			{Name: "color", Value: "clear"},
			{Name: "Color", Value: "matte"},
		}

		if err := s.Validate(FinishedProduct); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("delegates each spec to its own validate", func(t *testing.T) {
		s := Specs{{Name: "tensile", Value: "130 MPa"}}

		if err := s.Validate(RawMaterial); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})
}

func Test_Specs_Merge(t *testing.T) {
	t.Run("combines both sets in order", func(t *testing.T) {
		a := Specs{{Name: "thickness", Value: "20 micron"}}
		b := Specs{{Name: "tensile", Value: "130 MPa"}}

		merged := a.Merge(b)

		want := Specs{
			{Name: "thickness", Value: "20 micron"},
			{Name: "tensile", Value: "130 MPa"},
		}
		if !reflect.DeepEqual(merged, want) {
			t.Errorf("Merge() = %+v, want %+v", merged, want)
		}
	})

	t.Run("modifies neither operand", func(t *testing.T) {
		a := Specs{{Name: "thickness", Value: "20 micron"}}
		b := Specs{{Name: "tensile", Value: "130 MPa"}}

		_ = a.Merge(b)

		if len(a) != 1 || len(b) != 1 {
			t.Errorf("Merge() modified operands: %v, %v", a, b)
		}
	})
}

func Test_Specs_Columns(t *testing.T) {
	t.Run("splits names and values in order", func(t *testing.T) {
		s := Specs{
			{Name: "thickness", Value: "20 micron"},
			{Name: "tensile", Value: "130 MPa"},
		}

		names, values := s.Columns()

		if len(names) != 2 || names[0] != "thickness" || names[1] != "tensile" {
			t.Errorf("Columns() names = %v, want [thickness tensile]", names)
		}

		if len(values) != 2 || values[0] != "20 micron" || values[1] != "130 MPa" {
			t.Errorf("Columns() values = %v, want [20 micron 130 MPa]", values)
		}
	})

	t.Run("returns empty slices for an empty set", func(t *testing.T) {
		names, values := Specs(nil).Columns()

		if names == nil || values == nil {
			t.Error("Columns() returned nil, want empty slices")
		}

		if len(names) != 0 || len(values) != 0 {
			t.Errorf("Columns() = (%v, %v), want empty", names, values)
		}
	})
}
