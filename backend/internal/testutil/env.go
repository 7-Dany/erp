package testutil

import (
	"os"
	"testing"
)

// Unsetenv removes the variable for the test duration, then restores it.
func Unsetenv(t *testing.T, key string) {
	t.Helper()

	value, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) error = %v", key, err)
	}

	t.Cleanup(func() {
		if ok {
			t.Setenv(key, value)
		}
	})
}
