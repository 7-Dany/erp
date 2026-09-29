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

func Test_Query_Normalize(t *testing.T) {
	t.Run("trims the term and applies the default limit", func(t *testing.T) {
		q := Query{Term: "  bopp "}

		q.Normalize()

		if q.Term != "bopp" {
			t.Errorf("Term = %q, want %q", q.Term, "bopp")
		}

		if q.Limit != DefaultLimit {
			t.Errorf("Limit = %d, want default %d", q.Limit, DefaultLimit)
		}
	})

	t.Run("keeps an explicit limit", func(t *testing.T) {
		q := Query{Term: "bopp", Limit: 10}

		q.Normalize()

		if q.Limit != 10 {
			t.Errorf("Limit = %d, want 10", q.Limit)
		}
	})
}

func Test_Query_Validate(t *testing.T) {
	t.Run("accepts a normalized query", func(t *testing.T) {
		q := Query{Term: "bopp"}
		q.Normalize()

		if err := q.Validate(); err != nil {
			t.Fatalf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("rejects a blank term", func(t *testing.T) {
		for _, term := range []string{"", "   "} {
			q := Query{Term: term, Limit: DefaultLimit}

			if err := q.Validate(); !errors.Is(err, ErrInvalid) {
				t.Errorf("Validate() term %q error = %v, want ErrInvalid", term, err)
			}
		}
	})

	t.Run("rejects a term below the minimum length", func(t *testing.T) {
		q := Query{Term: "P", Limit: DefaultLimit}

		if err := q.Validate(); !errors.Is(err, ErrInvalid) {
			t.Fatalf("Validate() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects a limit outside the allowed range", func(t *testing.T) {
		for _, limit := range []int{-1, 1, 9, 51} {
			q := Query{Term: "bopp", Limit: limit}

			if err := q.Validate(); !errors.Is(err, ErrInvalid) {
				t.Errorf("Validate() limit %d error = %v, want ErrInvalid", limit, err)
			}
		}
	})
}
