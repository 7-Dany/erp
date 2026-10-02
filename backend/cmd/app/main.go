// main is the entrypoint for the application
package main

import (
	"context"
	"erp/internal/config"
	"erp/internal/db"
	"erp/internal/material"
	"fmt"
	"os"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.CreateDatabaseConfiguration(false)
	if err != nil {
		return err
	}

	database, err := config.NewDatabase(ctx, cfg)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := database.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}
	fmt.Println("Connected to PostgreSQL!")

	// The walkthrough writes to the real database; gate it behind ERP_DEMO=1.
	if os.Getenv("ERP_DEMO") != "1" {
		return nil
	}

	repository := material.NewPostgresRepository(db.New(database.Pool))

	created, err := repository.Create(ctx, material.Material{
		Name: "Demo Steel Sheet",
		Type: material.RawMaterial,
	})
	if err != nil {
		return err
	}
	fmt.Printf("--- create ---\n%+v\n", created)

	exists, err := repository.Exists(ctx, created.ID)
	if err != nil {
		return err
	}
	fmt.Printf("--- exists ---\nExists(%d): %v\n", created.ID, exists)

	stored, err := repository.Get(ctx, created.ID)
	if err != nil {
		return err
	}
	fmt.Printf("--- get ---\nGet(%d): %+v\n", created.ID, stored)

	for _, name := range []string{"Demo BOPP 20UM", "Demo BOPP Film", "Demo Copper Wire"} {
		if _, err := repository.Create(ctx, material.Material{
			Name: name,
			Type: material.RawMaterial,
		}); err != nil {
			return err
		}
	}

	matches, err := repository.Find(ctx, material.Query{Term: "bopp"})
	if err != nil {
		return err
	}
	fmt.Printf("--- find %q (%d matches) ---\n", "bopp", len(matches))
	for _, match := range matches {
		fmt.Printf("  %+v\n", match)
	}

	none, err := repository.Find(ctx, material.Query{Term: "no-such-demo-xyz"})
	if err != nil {
		return err
	}
	fmt.Printf("--- find %q (%d matches) ---\n", "no-such-demo-xyz", len(none))

	if _, err := repository.Find(ctx, material.Query{Term: " "}); err != nil {
		fmt.Printf("--- find blank ---\n%v\n", err)
	}

	withSpecs, err := repository.CreateWithSpecs(ctx, material.Material{
		Name: "Demo BOPP 20UM",
		Type: material.FinishedProduct,
		Specs: material.Specs{
			{Name: "thickness", Value: "20 micron"},
			{Name: "tensile", Value: "130 MPa"},
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("--- create with specs ---\n%+v\n", withSpecs)

	bare, err := repository.Create(ctx, material.Material{
		Name: "Demo PP Resin",
		Type: material.RawMaterial,
	})
	if err != nil {
		return err
	}

	if err := repository.AddSpecs(ctx, bare.ID, material.Specs{
		{Name: "grade", Value: "HOMO-25"},
		{Name: "mfi", Value: "3 g/10min"},
	}); err != nil {
		return err
	}

	full, err := repository.Get(ctx, bare.ID)
	if err != nil {
		return err
	}
	fmt.Printf("--- get with specs ---\nGet(%d): %+v\n", bare.ID, full)

	return nil
}
