package testutil

import (
	"erp/internal/config"
	"testing"
)

func Test_checkTestDatabase(t *testing.T) {
	normal := &config.DatabaseConfig{Host: "localhost", Port: "5432", Name: "business"}

	t.Run("accepts a _test database that differs from the normal one", func(t *testing.T) {
		test := config.DatabaseConfig{Host: "localhost", Port: "5432", Name: "business_test"}

		if err := checkTestDatabase(test, normal); err != nil {
			t.Fatalf("checkTestDatabase() error = %v, want nil", err)
		}
	})

	t.Run("accepts a _test database when the normal config is absent", func(t *testing.T) {
		test := config.DatabaseConfig{Host: "localhost", Port: "5432", Name: "business_test"}

		if err := checkTestDatabase(test, nil); err != nil {
			t.Fatalf("checkTestDatabase() error = %v, want nil", err)
		}
	})

	t.Run("rejects a name without the _test suffix", func(t *testing.T) {
		test := config.DatabaseConfig{Host: "localhost", Port: "5432", Name: "business_dev"}

		if err := checkTestDatabase(test, normal); err == nil {
			t.Fatal("checkTestDatabase() error = nil, want refusal")
		}
	})

	t.Run("rejects the same host, port and name as the normal database", func(t *testing.T) {
		same := &config.DatabaseConfig{Host: "localhost", Port: "5432", Name: "business_test"}
		test := config.DatabaseConfig{Host: "localhost", Port: "5432", Name: "business_test"}

		if err := checkTestDatabase(test, same); err == nil {
			t.Fatal("checkTestDatabase() error = nil, want refusal")
		}
	})
}
