package config_test

import (
	"context"
	"erp/internal/config"
	"testing"
)

func Test_NewDatabase(t *testing.T) {
	t.Run("creates a database from configuration", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:    "localhost",
			Port:    "5432",
			User:    "postgres",
			Name:    "business_test",
			SSLMode: "disable",
		}

		db, err := config.NewDatabase(context.Background(), cfg)
		if err != nil {
			t.Fatalf("NewDatabase() error = %v", err)
		}
		t.Cleanup(db.Close)

		if db.Config != cfg {
			t.Fatalf("Config = %+v, want %+v", db.Config, cfg)
		}
		if db.Pool == nil {
			t.Fatal("Pool = nil, want initialized pool")
		}
	})
	t.Run("returns an error for invalid configuration", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:    "local host with spaces",
			Port:    "not a port",
			User:    "postgres",
			Name:    "business_test",
			SSLMode: "disable",
		}

		if _, err := config.NewDatabase(context.Background(), cfg); err == nil {
			t.Fatal("NewDatabase() error = nil, want error")
		}
	})
}
