package material

import (
	"erp/internal/db"
	"errors"
	"strings"
	"testing"
)

func Test_Type_Valid(t *testing.T) {
	t.Run("accepts every defined type", func(t *testing.T) {
		defined := []Type{
			Unspecified, RawMaterial, Packaging,
			SemiFinished, FinishedProduct, SparePart,
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
			SemiFinished, FinishedProduct, SparePart,
		}

		for _, typ := range defined {
			if !db.MaterialType(typ).Valid() {
				t.Errorf("Type(%q) has no database enum value", typ)
			}
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
		err := Material{Name: "", Type: RawMaterial}.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a whitespace-only name", func(t *testing.T) {
		err := Material{Name: "  \t ", Type: RawMaterial}.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a name over the maximum length", func(t *testing.T) {
		name := strings.Repeat("a", MaxNameLength+1)

		err := Material{Name: name, Type: RawMaterial}.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects an unset type", func(t *testing.T) {
		err := Material{Name: "Steel Sheet"}.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects an unknown type", func(t *testing.T) {
		err := Material{Name: "Steel Sheet", Type: "liquid_gold"}.Validate()

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})
}
