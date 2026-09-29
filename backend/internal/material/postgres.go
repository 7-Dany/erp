package material

import (
	"context"
	"erp/internal/db"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// PostgresRepository is a PostgreSQL-backed Repository.
type PostgresRepository struct {
	queries db.Querier
}

// NewPostgresRepository backs a Repository with PostgreSQL.
func NewPostgresRepository(queries db.Querier) *PostgresRepository {
	return &PostgresRepository{
		queries: queries,
	}
}

// Create validates, trims, and stores a new material.
func (p *PostgresRepository) Create(ctx context.Context, material Material) (Material, error) {
	if err := material.Validate(); err != nil {
		return Material{}, fmt.Errorf("create material: %w", err)
	}

	params := db.CreateMaterialParams{
		Name: material.Name,
		Type: db.MaterialType(material.Type),
	}

	row, err := p.queries.CreateMaterial(ctx, params)
	if err != nil {
		return Material{}, fmt.Errorf("create material: %w", err)
	}

	return Material{
		ID:   row.ID,
		Name: row.Name,
		Type: Type(row.Type),
	}, nil
}

// Exists reports whether a material with the given ID exists.
func (p *PostgresRepository) Exists(ctx context.Context, id int64) (bool, error) {
	exists, err := p.queries.MaterialExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("check material exists: %w", err)
	}

	return exists, nil
}

// Get returns the material with the given ID, or ErrNotFound.
func (p *PostgresRepository) Get(ctx context.Context, id int64) (Material, error) {
	row, err := p.queries.GetMaterial(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Material{}, fmt.Errorf("get material %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return Material{}, fmt.Errorf("get material: %w", err)
	}

	return Material{
		ID:   row.ID,
		Name: row.Name,
		Type: Type(row.Type),
	}, nil
}

// Find returns materials whose name contains the query term, ordered by ID,
// up to the query limit. It never picks: every match is returned for the
// caller to choose from.
func (p *PostgresRepository) Find(ctx context.Context, query Query) ([]Material, error) {
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("find materials: %w", err)
	}

	rows, err := p.queries.FindMaterials(ctx, db.FindMaterialsParams{
		Query: query.EscapeTerm(),
		Limit: int64(query.Limit),
	})
	if err != nil {
		return nil, fmt.Errorf("find materials: %w", err)
	}

	materials := make([]Material, 0, len(rows))
	for _, row := range rows {
		materials = append(materials, Material{
			ID:   row.ID,
			Name: row.Name,
			Type: Type(row.Type),
		})
	}

	return materials, nil
}

var _ Repository = (*PostgresRepository)(nil)
