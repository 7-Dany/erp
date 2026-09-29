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
	fmt.Printf("Created: %+v\n", created)

	exists, err := repository.Exists(ctx, created.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Exists(%d): %v\n", created.ID, exists)

	stored, err := repository.Get(ctx, created.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Get(%d): %+v\n", created.ID, stored)

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
	fmt.Printf("Find(%q): %+v\n", "bopp", matches)

	none, err := repository.Find(ctx, material.Query{Term: "no-such-demo-xyz"})
	if err != nil {
		return err
	}
	fmt.Printf("Find(%q): %+v\n", "no-such-demo-xyz", none)

	if _, err := repository.Find(ctx, material.Query{Term: " "}); err != nil {
		fmt.Printf("Find(blank): %v\n", err)
	}

	return nil
}
