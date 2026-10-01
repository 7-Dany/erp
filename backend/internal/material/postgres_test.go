package material

import (
	"context"
	"erp/internal/testutil"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/puddle/v2"
)

func Test_PostgresRepository_Create(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials", "material_specs")
	t.Cleanup(pool.Close)

	t.Run("creates a material and returns a generated ID", func(t *testing.T) {
		input := Material{
			ID:   999,
			Name: "Polypropylene Film",
			Type: RawMaterial,
		}

		got, err := repository.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if got.ID == 0 {
			t.Fatal("Create() returned a zero ID")
		}

		if got.ID == input.ID {
			t.Fatalf(
				"Create() used caller-supplied ID %d; database should generate the ID",
				input.ID,
			)
		}

		if got.Name != input.Name {
			t.Errorf("Create() name = %q, want %q", got.Name, input.Name)
		}

		if got.Type != input.Type {
			t.Errorf("Create() type = %q, want %q", got.Type, input.Type)
		}
	})

	t.Run("different materials receive different IDs", func(t *testing.T) {
		first, err := repository.Create(ctx, Material{
			Name: "Steel Sheet",
			Type: RawMaterial,
		})
		if err != nil {
			t.Fatalf("Create() first material error = %v", err)
		}

		second, err := repository.Create(ctx, Material{
			Name: "Cardboard Box",
			Type: Packaging,
		})
		if err != nil {
			t.Fatalf("Create() second material error = %v", err)
		}

		if first.ID == second.ID {
			t.Fatalf(
				"two materials received the same ID: %d",
				first.ID,
			)
		}
	})

	t.Run("created material is actually persisted", func(t *testing.T) {
		input := Material{
			Name: "Finished Product",
			Type: FinishedProduct,
		}

		created, err := repository.Create(ctx, input)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		stored, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if !reflect.DeepEqual(stored, created) {
			t.Errorf(
				"persisted material = %+v, want %+v",
				stored,
				created,
			)
		}
	})

	t.Run("stores and returns each supported type", func(t *testing.T) {
		for _, typ := range []Type{
			Unspecified, RawMaterial, Packaging,
			SemiFinished, FinishedProduct, SparePart, Consumable,
		} {
			created, err := repository.Create(ctx, Material{
				Name: "Typed Material",
				Type: typ,
			})
			if err != nil {
				t.Fatalf("Create() type %q error = %v", typ, err)
			}

			if created.Type != typ {
				t.Errorf("Create() type = %q, want %q", created.Type, typ)
			}

			stored, err := repository.Get(ctx, created.ID)
			if err != nil {
				t.Fatalf("Get() type %q error = %v", typ, err)
			}

			if stored.Type != typ {
				t.Errorf("persisted type = %q, want %q", stored.Type, typ)
			}
		}
	})

	t.Run("rejects an invalid material", func(t *testing.T) {
		_, err := repository.Create(ctx, Material{
			Name: "   ",
			Type: RawMaterial,
		})

		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Create() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("trims the name before storing", func(t *testing.T) {
		created, err := repository.Create(ctx, Material{
			Name: "  Steel Sheet  ",
			Type: RawMaterial,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if created.Name != "Steel Sheet" {
			t.Fatalf("Create() name = %q, want %q", created.Name, "Steel Sheet")
		}

		stored, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if stored.Name != "Steel Sheet" {
			t.Errorf("persisted name = %q, want %q", stored.Name, "Steel Sheet")
		}
	})

	t.Run("returns an error when the database is unavailable", func(t *testing.T) {
		repository := testutil.NewClosedRepository(t, NewPostgresRepository)

		_, err := repository.Create(ctx, Material{
			Name: "Database Failure Test",
			Type: RawMaterial,
		})
		if !errors.Is(err, puddle.ErrClosedPool) {
			t.Fatalf("Create() error = %v, want closed pool", err)
		}
	})

	t.Run("rejects specs at creation", func(t *testing.T) {
		_, err := repository.Create(ctx, Material{
			Name:  "BOPP 20UM",
			Type:  FinishedProduct,
			Specs: Specs{{Name: "thickness", Value: "20 micron"}},
		})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("Create() error = %v, want ErrInvalid", err)
		}
	})
}

func Test_PostgresRepository_CreateWithSpecs(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials", "material_specs")
	t.Cleanup(pool.Close)

	t.Run("stores the material with its specs atomically", func(t *testing.T) {
		input := Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: Specs{
				{Name: "thickness", Value: "20 micron"},
				{Name: "tensile", Value: "130 MPa"},
			},
		}

		created, err := repository.CreateWithSpecs(ctx, input)
		if err != nil {
			t.Fatalf("CreateWithSpecs() error = %v", err)
		}

		if created.ID == 0 {
			t.Fatal("CreateWithSpecs() returned a zero ID")
		}

		stored, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		want := Specs{
			{Name: "tensile", Value: "130 MPa"},
			{Name: "thickness", Value: "20 micron"},
		}
		if !reflect.DeepEqual(stored.Specs, want) {
			t.Errorf("Get() specs = %+v, want %+v in name order", stored.Specs, want)
		}
	})

	t.Run("stores a material without specs", func(t *testing.T) {
		created, err := repository.CreateWithSpecs(ctx, Material{
			Name: "PP Resin",
			Type: RawMaterial,
		})
		if err != nil {
			t.Fatalf("CreateWithSpecs() error = %v", err)
		}

		stored, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if len(stored.Specs) != 0 {
			t.Errorf("Get() specs = %+v, want empty", stored.Specs)
		}
	})

	t.Run("rejects invalid specs without storing anything", func(t *testing.T) {
		_, err := repository.CreateWithSpecs(ctx, Material{
			Name: "PP Resin",
			Type: RawMaterial,
			Specs: Specs{
				{Name: "tensile", Value: "130 MPa"},
			},
		})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("CreateWithSpecs() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("returns an error when the database is unavailable", func(t *testing.T) {
		repository := testutil.NewClosedRepository(t, NewPostgresRepository)

		_, err := repository.CreateWithSpecs(ctx, Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
			Specs: Specs{
				{Name: "thickness", Value: "20 micron"},
			},
		})
		if !errors.Is(err, puddle.ErrClosedPool) {
			t.Fatalf("CreateWithSpecs() error = %v, want closed pool", err)
		}
	})
}

func Test_PostgresRepository_AddSpecs(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials", "material_specs")
	t.Cleanup(pool.Close)

	t.Run("attaches specs retrievable with the material", func(t *testing.T) {
		created, err := repository.Create(ctx, Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		specs := Specs{
			{Name: "thickness", Value: "20 micron"},
			{Name: "tensile", Value: "130 MPa"},
		}
		if err := repository.AddSpecs(ctx, created.ID, specs); err != nil {
			t.Fatalf("AddSpecs() error = %v", err)
		}

		stored, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		want := Specs{
			{Name: "tensile", Value: "130 MPa"},
			{Name: "thickness", Value: "20 micron"},
		}
		if !reflect.DeepEqual(stored.Specs, want) {
			t.Errorf("Get() specs = %+v, want %+v in name order", stored.Specs, want)
		}
	})

	t.Run("accepts an empty set as a no-op", func(t *testing.T) {
		created, err := repository.Create(ctx, Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if err := repository.AddSpecs(ctx, created.ID, nil); err != nil {
			t.Fatalf("AddSpecs() error = %v, want nil", err)
		}

		stored, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if len(stored.Specs) != 0 {
			t.Errorf("Get() specs = %+v, want empty", stored.Specs)
		}
	})

	t.Run("rejects specs outside the type profile", func(t *testing.T) {
		created, err := repository.Create(ctx, Material{
			Name: "PP Resin",
			Type: RawMaterial,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		err = repository.AddSpecs(ctx, created.ID, Specs{{Name: "tensile", Value: "130 MPa"}})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("AddSpecs() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("rejects specs duplicating stored ones", func(t *testing.T) {
		created, err := repository.Create(ctx, Material{
			Name: "BOPP 20UM",
			Type: FinishedProduct,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if err := repository.AddSpecs(ctx, created.ID, Specs{{Name: "color", Value: "clear"}}); err != nil {
			t.Fatalf("AddSpecs() error = %v", err)
		}

		err = repository.AddSpecs(ctx, created.ID, Specs{{Name: "Color", Value: "matte"}})
		if !errors.Is(err, ErrInvalid) {
			t.Fatalf("AddSpecs() error = %v, want ErrInvalid", err)
		}
	})

	t.Run("returns ErrNotFound for a missing material", func(t *testing.T) {
		err := repository.AddSpecs(ctx, 999999999, Specs{{Name: "thickness", Value: "20 micron"}})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("AddSpecs() error = %v, want ErrNotFound", err)
		}
	})

	t.Run("returns an error when the database is unavailable", func(t *testing.T) {
		repository := testutil.NewClosedRepository(t, NewPostgresRepository)

		err := repository.AddSpecs(ctx, 1, Specs{{Name: "thickness", Value: "20 micron"}})
		if !errors.Is(err, puddle.ErrClosedPool) {
			t.Fatalf("AddSpecs() error = %v, want closed pool", err)
		}
	})
}

func Test_PostgresRepository_Exists(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials", "material_specs")
	t.Cleanup(pool.Close)

	t.Run("returns true for an existing material", func(t *testing.T) {
		created, err := repository.Create(ctx, Material{
			Name: "Existing Material",
			Type: RawMaterial,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		exists, err := repository.Exists(ctx, created.ID)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}

		if !exists {
			t.Fatal("Exists() = false, want true")
		}
	})

	t.Run("returns false for a material that does not exist", func(t *testing.T) {
		exists, err := repository.Exists(ctx, 999999999)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}

		if exists {
			t.Fatal("Exists() = true, want false")
		}
	})

	t.Run("returns an error when the database is unavailable", func(t *testing.T) {
		repository := testutil.NewClosedRepository(t, NewPostgresRepository)

		_, err := repository.Exists(ctx, 1)
		if !errors.Is(err, puddle.ErrClosedPool) {
			t.Fatalf("Exists() error = %v, want closed pool", err)
		}
	})
}

func Test_PostgresRepository_Get(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials", "material_specs")
	t.Cleanup(pool.Close)

	t.Run("retrieves an existing material", func(t *testing.T) {
		created, err := repository.Create(ctx, Material{
			Name: "Retrieved Material",
			Type: SemiFinished,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}

		if !reflect.DeepEqual(got, created) {
			t.Errorf(
				"Get() = %+v, want %+v",
				got,
				created,
			)
		}
	})

	t.Run("returns ErrNotFound when the material does not exist", func(t *testing.T) {
		_, err := repository.Get(ctx, 999999999)
		if err == nil {
			t.Fatal("Get() error = nil, want ErrNotFound")
		}

		if !errors.Is(err, ErrNotFound) {
			t.Fatalf(
				"Get() error = %v, want ErrNotFound",
				err,
			)
		}
	})

	t.Run("returns an error when the database is unavailable", func(t *testing.T) {
		repository := testutil.NewClosedRepository(t, NewPostgresRepository)

		_, err := repository.Get(ctx, 1)
		if !errors.Is(err, puddle.ErrClosedPool) {
			t.Fatalf("Get() error = %v, want closed pool", err)
		}
	})
}

func Test_PostgresRepository_Find(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials", "material_specs")
	t.Cleanup(pool.Close)

	seed := func(t *testing.T, name string) Material {
		t.Helper()

		created, err := repository.Create(ctx, Material{
			Name: name,
			Type: RawMaterial,
		})
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}

		return created
	}

	t.Run("returns each match once ordered by ID with ID name and type", func(t *testing.T) {
		first := seed(t, "BOPP 20UM")

		second, err := repository.Create(ctx, Material{
			Name: "BOPP FILM 20 MICRON",
			Type: Packaging,
		})
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		seed(t, "unrelated coating")

		got, err := repository.Find(ctx, Query{Term: "BOPP", Limit: 0})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if len(got) != 2 {
			t.Fatalf("Find() returned %d materials, want 2", len(got))
		}

		if !reflect.DeepEqual(got, []Material{first, second}) {
			t.Errorf("Find() = %+v, want [%+v %+v]", got, first, second)
		}
	})

	t.Run("returns an empty list when nothing matches", func(t *testing.T) {
		got, err := repository.Find(ctx, Query{Term: "no such material xyz", Limit: 0})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if len(got) != 0 {
			t.Fatalf("Find() returned %+v, want empty", got)
		}
	})

	t.Run("rejects blank and too-short entries", func(t *testing.T) {
		for _, query := range []string{"", "   ", "P"} {
			_, err := repository.Find(ctx, Query{Term: query})
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("Find(%q) error = %v, want ErrInvalid", query, err)
			}
		}
	})

	t.Run("accepts a two-character entry", func(t *testing.T) {
		first := seed(t, "aluminum rod")
		second := seed(t, "aluminum tube")

		got, err := repository.Find(ctx, Query{Term: "al", Limit: 0})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if !reflect.DeepEqual(got, []Material{first, second}) {
			t.Fatalf("Find() = %+v, want [%+v %+v]", got, first, second)
		}
	})

	t.Run("shows the first limit rows and validates the limit", func(t *testing.T) {
		for _, limit := range []int{-100, -1, 1, 9, 51, 1000} {
			_, err := repository.Find(ctx, Query{Term: "bulk film", Limit: limit})
			if !errors.Is(err, ErrInvalid) {
				t.Errorf("Find(limit=%d) error = %v, want ErrInvalid", limit, err)
			}
		}

		if DefaultLimit != 20 {
			t.Fatalf("DefaultLimit = %d, want 20", DefaultLimit)
		}

		var created []Material
		for i := 0; i < 21; i++ {
			created = append(created, seed(t, "bulk film batch"))
		}

		got, err := repository.Find(ctx, Query{Term: "bulk film", Limit: 10})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if len(got) != 10 {
			t.Fatalf("Find() returned %d materials, want 10", len(got))
		}

		for i, want := range created[:10] {
			if !reflect.DeepEqual(got[i], want) {
				t.Fatalf("Find()[%d] = %+v, want %+v", i, got[i], want)
			}
		}

		got, err = repository.Find(ctx, Query{Term: "bulk film", Limit: 0})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if len(got) != 20 {
			t.Fatalf("Find() returned %d materials, want default 20", len(got))
		}

		var ranged []Material
		for i := 0; i < 50; i++ {
			ranged = append(ranged, seed(t, "limit range spool"))
		}

		got, err = repository.Find(ctx, Query{Term: "limit range spool", Limit: 50})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if len(got) != 50 {
			t.Fatalf("Find() returned %d materials, want upper boundary 50", len(got))
		}

		for i, want := range ranged {
			if !reflect.DeepEqual(got[i], want) {
				t.Fatalf("Find()[%d] = %+v, want %+v", i, got[i], want)
			}
		}
	})

	t.Run("returns an error when the database is unavailable", func(t *testing.T) {
		repository := testutil.NewClosedRepository(t, NewPostgresRepository)

		_, err := repository.Find(ctx, Query{Term: "BOPP", Limit: 0})
		if !errors.Is(err, puddle.ErrClosedPool) {
			t.Fatalf("Find() error = %v, want closed pool", err)
		}
	})

	t.Run("trims the entry and matches case-insensitively", func(t *testing.T) {
		want := seed(t, "BOPP 20UM grade")

		got, err := repository.Find(ctx, Query{Term: "  bopp 20um grade ", Limit: 0})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if !reflect.DeepEqual(got, []Material{want}) {
			t.Fatalf("Find() = %+v, want [%+v]", got, want)
		}
	})

	t.Run("shows similar names without picking one", func(t *testing.T) {
		first := seed(t, "polypropylene HOMO")
		second := seed(t, "polypropylene COPO")

		got, err := repository.Find(ctx, Query{Term: "polypropylene", Limit: 0})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if !reflect.DeepEqual(got, []Material{first, second}) {
			t.Fatalf("Find() = %+v, want both matches unpicked", got)
		}
	})

	t.Run("returns the matches for the employee to confirm against", func(t *testing.T) {
		first := seed(t, "steel sheet")
		second := seed(t, "steel strip")

		got, err := repository.Find(ctx, Query{Term: "steel", Limit: 0})
		if err != nil {
			t.Fatalf("Find() error = %v", err)
		}

		if !reflect.DeepEqual(got, []Material{first, second}) {
			t.Fatalf("Find() = %+v, want both matches for confirm step", got)
		}
	})
}
