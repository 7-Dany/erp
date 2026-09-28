package config_test

import (
	"erp/internal/config"
	"erp/internal/testutil"
	"testing"
)

func Test_DatabaseConfig_URL(t *testing.T) {
	t.Run("assembles URL from parts", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "secret",
			Name:     "business_test",
			SSLMode:  "disable",
		}

		want := "postgres://postgres:secret@localhost:5432/business_test?sslmode=disable" //nolint:gosec
		if got := cfg.URL(); got != want {
			t.Fatalf("URL() = %q, want %q", got, want)
		}
	})

	t.Run("omits password when empty", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:    "localhost",
			Port:    "5432",
			User:    "postgres",
			Name:    "business_test",
			SSLMode: "disable",
		}

		want := "postgres://postgres@localhost:5432/business_test?sslmode=disable"
		if got := cfg.URL(); got != want {
			t.Fatalf("URL() = %q, want %q", got, want)
		}
	})

	t.Run("escapes special characters in password", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "p@ss:word",
			Name:     "business_test",
			SSLMode:  "disable",
		}

		want := "postgres://postgres:p%40ss%3Aword@localhost:5432/business_test?sslmode=disable" //nolint:gosec
		if got := cfg.URL(); got != want {
			t.Fatalf("URL() = %q, want %q", got, want)
		}
	})
}

func Test_DatabaseConfig_SameDatabase(t *testing.T) {
	base := config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "secret",
		Name:     "business",
		SSLMode:  "disable",
	}

	t.Run("is true for identical configs", func(t *testing.T) {
		if !base.SameDatabase(base) {
			t.Fatal("SameDatabase() = false, want true")
		}
	})

	t.Run("is true when only credentials and SSL mode differ", func(t *testing.T) {
		other := base
		other.User = "dany"
		other.Password = "other"
		other.SSLMode = "require"

		if !base.SameDatabase(other) {
			t.Fatal("SameDatabase() = false, want true")
		}
	})

	t.Run("is false when the host differs", func(t *testing.T) {
		other := base
		other.Host = "db.internal"

		if base.SameDatabase(other) {
			t.Fatal("SameDatabase() = true, want false")
		}
	})

	t.Run("is false when the port differs", func(t *testing.T) {
		other := base
		other.Port = "5433"

		if base.SameDatabase(other) {
			t.Fatal("SameDatabase() = true, want false")
		}
	})

	t.Run("is false when the name differs", func(t *testing.T) {
		other := base
		other.Name = "business_test"

		if base.SameDatabase(other) {
			t.Fatal("SameDatabase() = true, want false")
		}
	})
}

func Test_SSLMode_Valid(t *testing.T) {
	t.Run("accepts every supported mode", func(t *testing.T) {
		modes := []config.SSLMode{
			config.SSLDisable, config.SSLAllow, config.SSLPrefer,
			config.SSLRequire, config.SSLVerifyCA, config.SSLVerifyFull,
		}

		for _, mode := range modes {
			if !mode.Valid() {
				t.Errorf("SSLMode(%q).Valid() = false, want true", mode)
			}
		}
	})

	t.Run("rejects unknown modes", func(t *testing.T) {
		for _, mode := range []config.SSLMode{"", "required", "DISABLE"} {
			if mode.Valid() {
				t.Errorf("SSLMode(%q).Valid() = true, want false", mode)
			}
		}
	})
}

func Test_CreateDatabaseConfiguration(t *testing.T) {
	t.Run("loads test parts with defaults", func(t *testing.T) {
		t.Setenv("TEST_DB_USER", "postgres")
		t.Setenv("TEST_DB_PASSWORD", "secret")
		t.Setenv("TEST_DB_NAME", "business_test")
		testutil.Unsetenv(t, "TEST_DB_SSLMODE")

		cfg, err := config.CreateDatabaseConfiguration(true)
		if err != nil {
			t.Fatalf("CreateDatabaseConfiguration() error = %v", err)
		}

		if cfg.Host != "localhost" {
			t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
		}
		if cfg.Port != "5432" {
			t.Errorf("Port = %q, want %q", cfg.Port, "5432")
		}
		if cfg.SSLMode != "disable" {
			t.Errorf("SSLMode = %q, want %q", cfg.SSLMode, "disable")
		}
		if cfg.User != "postgres" || cfg.Password != "secret" || cfg.Name != "business_test" {
			t.Errorf("loaded config = %+v", cfg)
		}
	})

	t.Run("loads app parts", func(t *testing.T) {
		t.Setenv("DB_USER", "dany")
		t.Setenv("DB_NAME", "business")
		t.Setenv("DB_SSLMODE", "require")

		cfg, err := config.CreateDatabaseConfiguration(false)
		if err != nil {
			t.Fatalf("CreateDatabaseConfiguration() error = %v", err)
		}
		if cfg.User != "dany" || cfg.Name != "business" {
			t.Errorf("loaded config = %+v", cfg)
		}
		if cfg.SSLMode != "require" {
			t.Errorf("SSLMode = %q, want %q", cfg.SSLMode, "require")
		}
	})

	t.Run("returns error when app sslmode is missing", func(t *testing.T) {
		t.Setenv("DB_USER", "dany")
		t.Setenv("DB_NAME", "business")
		testutil.Unsetenv(t, "DB_SSLMODE")

		if _, err := config.CreateDatabaseConfiguration(false); err == nil {
			t.Fatal("CreateDatabaseConfiguration() error = nil, want error")
		}
	})

	t.Run("returns error when app sslmode is blank", func(t *testing.T) {
		t.Setenv("DB_USER", "dany")
		t.Setenv("DB_NAME", "business")
		t.Setenv("DB_SSLMODE", "  ")

		if _, err := config.CreateDatabaseConfiguration(false); err == nil {
			t.Fatal("CreateDatabaseConfiguration() error = nil, want error")
		}
	})

	t.Run("returns error when test sslmode is invalid", func(t *testing.T) {
		t.Setenv("TEST_DB_USER", "postgres")
		t.Setenv("TEST_DB_NAME", "business_test")
		t.Setenv("TEST_DB_SSLMODE", "sometimes")

		if _, err := config.CreateDatabaseConfiguration(true); err == nil {
			t.Fatal("CreateDatabaseConfiguration() error = nil, want error")
		}
	})

	t.Run("returns error when app sslmode is invalid", func(t *testing.T) {
		t.Setenv("DB_USER", "dany")
		t.Setenv("DB_NAME", "business")
		t.Setenv("DB_SSLMODE", "sometimes")

		if _, err := config.CreateDatabaseConfiguration(false); err == nil {
			t.Fatal("CreateDatabaseConfiguration() error = nil, want error")
		}
	})

	t.Run("returns error when user is missing", func(t *testing.T) {
		testutil.Unsetenv(t, "TEST_DB_USER")
		t.Setenv("TEST_DB_NAME", "business_test")

		if _, err := config.CreateDatabaseConfiguration(true); err == nil {
			t.Fatal("CreateDatabaseConfiguration() error = nil, want error")
		}
	})

	t.Run("returns error when name is missing", func(t *testing.T) {
		testutil.Unsetenv(t, "TEST_DB_NAME")
		t.Setenv("TEST_DB_USER", "postgres")

		if _, err := config.CreateDatabaseConfiguration(true); err == nil {
			t.Fatal("CreateDatabaseConfiguration() error = nil, want error")
		}
	})
}
