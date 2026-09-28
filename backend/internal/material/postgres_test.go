package material

import (
	"context"
	"erp/internal/testutil"
	"errors"
	"testing"

	"github.com/jackc/puddle/v2"
)

func Test_PostgresRepository_Create(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials")
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

		if stored != created {
			t.Errorf(
				"persisted material = %+v, want %+v",
				stored,
				created,
			)
		}
	})

	// Alt path 1.1: validation rules are covered in Test_Material_Validate;
	// this checks Create enforces them.
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
}

func Test_PostgresRepository_Exists(t *testing.T) {
	ctx := context.Background()

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials")
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

	repository, pool := testutil.NewRepository(t, NewPostgresRepository, "materials")
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

		if got != created {
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
