-- name: CreateMaterial :one
INSERT INTO materials (
    name,
    type
)
VALUES (
    $1,
    $2
)
RETURNING id, name, type;

-- name: MaterialExists :one
SELECT EXISTS (
    SELECT 1
    FROM materials
    WHERE id = $1
);

-- name: GetMaterial :one
SELECT
    id,
    name,
    type
FROM materials
WHERE id = $1;
