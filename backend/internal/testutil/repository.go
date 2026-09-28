package testutil

import (
	"erp/internal/db"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRepository migrates, truncates tables, and builds a repository
// with build, returning it with its test-owned pool.
func NewRepository[R any](t *testing.T, build func(db.Querier) R, tables ...string) (R, *pgxpool.Pool) {
	t.Helper()

	pool := NewPool(t, tables...)

	return build(db.New(pool)), pool
}

// NewClosedRepository builds a repository over an already-closed pool.
func NewClosedRepository[R any](t *testing.T, build func(db.Querier) R) R {
	t.Helper()

	return build(db.New(ClosedPool(t)))
}
