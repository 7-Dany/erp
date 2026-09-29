package material

import "context"

// Repository persists materials.
type Repository interface {
	Create(ctx context.Context, material Material) (Material, error)
	Exists(ctx context.Context, id int64) (bool, error)
	Get(ctx context.Context, id int64) (Material, error)
	Find(ctx context.Context, query Query) ([]Material, error)
}
