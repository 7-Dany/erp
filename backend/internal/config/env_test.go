package config

import (
	"testing"
)

func Test_Get(t *testing.T) {
	t.Run("returns fallback when variable is unset", func(t *testing.T) {
		if got := Get("ERP_TEST_GET_MISSING_FALLBACK", "fallback"); got != "fallback" {
			t.Fatalf("Get() = %q, want %q", got, "fallback")
		}
	})

	t.Run("returns variable value when set", func(t *testing.T) {
		t.Setenv("ERP_TEST_GET_SET", "value")

		if got := Get("ERP_TEST_GET_SET", "fallback"); got != "value" {
			t.Fatalf("Get() = %q, want %q", got, "value")
		}
	})

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		t.Setenv("ERP_TEST_GET_TRIM", "  value  ")

		if got := Get("ERP_TEST_GET_TRIM", "fallback"); got != "value" {
			t.Fatalf("Get() = %q, want %q", got, "value")
		}
	})

	t.Run("returns fallback when value is blank", func(t *testing.T) {
		t.Setenv("ERP_TEST_GET_BLANK", "   ")

		if got := Get("ERP_TEST_GET_BLANK", "fallback"); got != "fallback" {
			t.Fatalf("Get() = %q, want %q", got, "fallback")
		}
	})
}

func Test_Lookup(t *testing.T) {
	t.Run("returns value and true when set", func(t *testing.T) {
		t.Setenv("ERP_TEST_LOOKUP_SET", "  value  ")

		got, ok := Lookup("ERP_TEST_LOOKUP_SET")
		if !ok || got != "value" {
			t.Fatalf("Lookup() = (%q, %v), want (%q, true)", got, ok, "value")
		}
	})

	t.Run("returns false when variable is unset", func(t *testing.T) {
		if _, ok := Lookup("ERP_TEST_LOOKUP_MISSING"); ok {
			t.Fatal("Lookup() ok = true, want false")
		}
	})

	t.Run("returns false when value is blank", func(t *testing.T) {
		t.Setenv("ERP_TEST_LOOKUP_BLANK", "   ")

		if _, ok := Lookup("ERP_TEST_LOOKUP_BLANK"); ok {
			t.Fatal("Lookup() ok = true, want false")
		}
	})
}

func Test_Required(t *testing.T) {
	t.Run("returns trimmed value when set", func(t *testing.T) {
		t.Setenv("ERP_TEST_REQUIRED_SET", "  value  ")

		got, err := Required("ERP_TEST_REQUIRED_SET")
		if err != nil {
			t.Fatalf("Required() error = %v", err)
		}
		if got != "value" {
			t.Fatalf("Required() = %q, want %q", got, "value")
		}
	})

	t.Run("returns error when variable is unset", func(t *testing.T) {
		if _, err := Required("ERP_TEST_REQUIRED_MISSING"); err == nil {
			t.Fatal("Required() error = nil, want error")
		}
	})

	t.Run("returns error when value is blank", func(t *testing.T) {
		t.Setenv("ERP_TEST_REQUIRED_BLANK", "   ")

		if _, err := Required("ERP_TEST_REQUIRED_BLANK"); err == nil {
			t.Fatal("Required() error = nil, want error")
		}
	})
}
