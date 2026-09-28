package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database is a PostgreSQL connection pool.
type Database struct {
	Config DatabaseConfig
	Pool   *pgxpool.Pool
}

// NewDatabase opens a pool from cfg.
func NewDatabase(ctx context.Context, cfg DatabaseConfig) (*Database, error) {
	pool, err := pgxpool.New(ctx, cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	return &Database{Config: cfg, Pool: pool}, nil
}

// Close releases the connection pool.
func (d *Database) Close() {
	d.Pool.Close()
}
