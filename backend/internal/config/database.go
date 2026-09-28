package config

import (
	"fmt"
	"net/url"
)

// SSLMode is a PostgreSQL sslmode.
type SSLMode string

// Supported PostgreSQL sslmodes.
const (
	SSLDisable    SSLMode = "disable"
	SSLAllow      SSLMode = "allow"
	SSLPrefer     SSLMode = "prefer"
	SSLRequire    SSLMode = "require"
	SSLVerifyCA   SSLMode = "verify-ca"
	SSLVerifyFull SSLMode = "verify-full"
)

// Valid reports whether m is a supported sslmode.
func (m SSLMode) Valid() bool {
	switch m {
	case SSLDisable, SSLAllow, SSLPrefer, SSLRequire, SSLVerifyCA, SSLVerifyFull:
		return true
	}

	return false
}

// DatabaseConfig holds PostgreSQL connection parts.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  SSLMode
}

// URL builds the PostgreSQL connection string.
func (c DatabaseConfig) URL() string {
	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   "/" + c.Name,
	}
	if c.Password == "" {
		u.User = url.User(c.User)
	} else {
		u.User = url.UserPassword(c.User, c.Password)
	}

	query := u.Query()
	query.Set("sslmode", string(c.SSLMode))
	u.RawQuery = query.Encode()

	return u.String()
}

// SameDatabase reports whether c and other address the same database,
// ignoring credentials and SSL mode.
func (c DatabaseConfig) SameDatabase(other DatabaseConfig) bool {
	return c.Host == other.Host && c.Port == other.Port && c.Name == other.Name
}

// CreateDatabaseConfiguration loads TEST_DB_* when isTest is true, DB_*
// otherwise. SSL mode is required for the app config, defaulting to
// disable for tests.
func CreateDatabaseConfiguration(isTest bool) (DatabaseConfig, error) {
	prefix := "DB"
	if isTest {
		prefix = "TEST_DB"
	}

	user, err := Required(prefix + "_USER")
	if err != nil {
		return DatabaseConfig{}, err
	}

	name, err := Required(prefix + "_NAME")
	if err != nil {
		return DatabaseConfig{}, err
	}

	// The test config defaults to disable; the app config requires it.
	rawMode, ok := Lookup(prefix + "_SSLMODE")
	if !ok {
		if !isTest {
			return DatabaseConfig{}, fmt.Errorf("%s is not set", prefix+"_SSLMODE")
		}
		rawMode = string(SSLDisable)
	}

	sslMode := SSLMode(rawMode)
	if !sslMode.Valid() {
		return DatabaseConfig{}, fmt.Errorf("unknown sslmode %q", rawMode)
	}

	return DatabaseConfig{
		Host:     Get(prefix+"_HOST", "localhost"),
		Port:     Get(prefix+"_PORT", "5432"),
		User:     user,
		Password: Get(prefix+"_PASSWORD", ""),
		Name:     name,
		SSLMode:  sslMode,
	}, nil
}
