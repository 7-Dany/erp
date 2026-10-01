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

// Create validates and stores a new material without specs.
func (p *PostgresRepository) Create(ctx context.Context, material Material) (Material, error) {
	if len(material.Specs) > 0 {
		return Material{}, fmt.Errorf(
			"create material: %w: use CreateWithSpecs",
			ErrInvalid,
		)
	}

	if err := material.Validate(); err != nil {
		return Material{}, fmt.Errorf("create material: %w", err)
	}

	row, err := p.queries.CreateMaterial(ctx, db.CreateMaterialParams{
		Name: material.Name,
		Type: db.MaterialType(material.Type),
	})
	if err != nil {
		return Material{}, fmt.Errorf("create material: %w", err)
	}

	material.ID = row.ID

	return material, nil
}

// CreateWithSpecs validates and stores a new material with its specs
// in one atomic statement.
func (p *PostgresRepository) CreateWithSpecs(ctx context.Context, material Material) (Material, error) {
	if err := material.Validate(); err != nil {
		return Material{}, fmt.Errorf("create material: %w", err)
	}

	names, values := material.Specs.Columns()

	row, err := p.queries.CreateMaterialWithSpecs(ctx, db.CreateMaterialWithSpecsParams{
		Name:       material.Name,
		Type:       db.MaterialType(material.Type),
		Names:      names,
		SpecValues: values,
	})
	if err != nil {
		return Material{}, fmt.Errorf("create material: %w", err)
	}

	material.ID = row.ID

	return material, nil
}

// AddSpecs validates and attaches specs to an existing material.
func (p *PostgresRepository) AddSpecs(ctx context.Context, id int64, specs Specs) error {
	stored, err := p.Get(ctx, id)
	if err != nil {
		return err
	}

	merged := stored.Specs.Merge(specs)

	temp := Material{ID: id, Name: stored.Name, Type: stored.Type, Specs: merged}
	if err := temp.Validate(); err != nil {
		return fmt.Errorf("add specs to material %d: %w", id, err)
	}

	names, values := specs.Columns()

	if err := p.queries.AddMaterialSpecs(ctx, db.AddMaterialSpecsParams{
		MaterialID: id,
		Names:      names,
		SpecValues: values,
	}); err != nil {
		return fmt.Errorf("add specs to material %d: %w", id, err)
	}

	return nil
}

// Exists reports whether a material with the given ID exists.
func (p *PostgresRepository) Exists(ctx context.Context, id int64) (bool, error) {
	exists, err := p.queries.MaterialExists(ctx, id)
	if err != nil {
		return false, fmt.Errorf("check material exists: %w", err)
	}

	return exists, nil
}

// Get returns the material with the given ID, with its specs, or ErrNotFound.
func (p *PostgresRepository) Get(ctx context.Context, id int64) (Material, error) {
	row, err := p.queries.GetMaterial(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Material{}, fmt.Errorf("get material %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return Material{}, fmt.Errorf("get material: %w", err)
	}

	specRows, err := p.queries.GetMaterialSpecs(ctx, id)
	if err != nil {
		return Material{}, fmt.Errorf("get material %d specs: %w", id, err)
	}

	material := Material{
		ID:   row.ID,
		Name: row.Name,
		Type: Type(row.Type),
	}
	material.AttachSpecs(specRows)

	return material, nil
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
