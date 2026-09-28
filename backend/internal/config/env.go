// Package config loads application configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Lookup returns the trimmed value of key and whether it is set and non-blank.
func Lookup(key string) (string, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	return value, value != ""
}

// Get returns the trimmed value of key, or fallback when unset or blank.
func Get(key, fallback string) string {
	if value, ok := Lookup(key); ok {
		return value
	}

	return fallback
}

// Required returns the trimmed value of key, or an error when unset or blank.
func Required(key string) (string, error) {
	if value, ok := Lookup(key); ok {
		return value, nil
	}

	return "", fmt.Errorf("%s is not set", key)
}
