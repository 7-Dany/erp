// Package testutil provides shared helpers for integration tests.
package testutil

import (
	"context"
	"database/sql"
	"erp/internal/config"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	// Register the pgx driver under database/sql for goose migrations.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// NewPostgres migrates the test database and returns a pool owned by the test.
func NewPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping PostgreSQL integration test in short mode")
	}

	cfg, err := config.CreateDatabaseConfiguration(true)
	if err != nil {
		t.Fatal(err)
	}
	var normal *config.DatabaseConfig
	if app, appErr := config.CreateDatabaseConfiguration(false); appErr == nil {
		normal = &app
	}
	if err = checkTestDatabase(cfg, normal); err != nil {
		t.Fatal(err)
	}

	databaseURL := cfg.URL()

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}
	if err := goose.Up(sqlDB, migrationsDir(t)); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	db, err := config.NewDatabase(context.Background(), cfg)
	if err != nil {
		t.Fatalf("create PostgreSQL database: %v", err)
	}
	t.Cleanup(db.Close)

	return db.Pool
}

// NewPool migrates, truncates tables, and returns a test-owned pool.
func NewPool(t *testing.T, tables ...string) *pgxpool.Pool {
	t.Helper()

	pool := NewPostgres(t)
	if len(tables) > 0 {
		Truncate(t, pool, tables...)
	}

	return pool
}

// ClosedPool returns an already-closed pool.
func ClosedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool := NewPostgres(t)
	pool.Close()

	return pool
}

// Truncate empties tables and restarts identities.
func Truncate(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()

	if len(tables) == 0 {
		t.Fatal("Truncate requires at least one table")
	}

	// Compare server-side: localhost and 127.0.0.1 reach the same database.
	var name string
	if err := pool.QueryRow(context.Background(), "SELECT current_database()").Scan(&name); err != nil {
		t.Fatalf("read current database: %v", err)
	}
	if !strings.HasSuffix(name, testDatabaseSuffix) {
		t.Fatalf("refusing to truncate %q: name must end in %q", name, testDatabaseSuffix)
	}

	quoted := make([]string, len(tables))
	for i, table := range tables {
		quoted[i] = `"` + strings.ReplaceAll(table, `"`, `""`) + `"`
	}

	_, err := pool.Exec(
		context.Background(),
		"TRUNCATE "+strings.Join(quoted, ", ")+" RESTART IDENTITY CASCADE",
	)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

const testDatabaseSuffix = "_test"

// checkTestDatabase refuses destructive setup against a non-test database.
// normal is nil when DB_* is unconfigured (e.g. CI).
func checkTestDatabase(test config.DatabaseConfig, normal *config.DatabaseConfig) error {
	if !strings.HasSuffix(test.Name, testDatabaseSuffix) {
		return fmt.Errorf("refusing to run: TEST_DB_NAME %q must end in %q", test.Name, testDatabaseSuffix)
	}

	if normal != nil && test.SameDatabase(*normal) {
		return fmt.Errorf("refusing to run: TEST_DB_* points at the same database as DB_*")
	}

	return nil
}

// migrationsDir finds sql/schema regardless of working directory.
func migrationsDir(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine testutil file path")
	}

	return filepath.Clean(
		filepath.Join(filepath.Dir(currentFile), "../../sql/schema"),
	)
}
