-- name: CreateMaterial :one
INSERT INTO materials (
    name,
    type
)
VALUES (
    $1,
    $2
)
RETURNING
    id,
    name,
    type;

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

-- name: FindMaterials :many
SELECT
    id,
    name,
    type
FROM materials
WHERE name ILIKE '%' || sqlc.arg(query)::text || '%' ESCAPE '\'
ORDER BY id
LIMIT $2;

-- name: AddMaterialSpecs :exec
INSERT INTO material_specs (
    material_id,
    name,
    value
)
SELECT
    sqlc.arg(material_id),
    unnest(sqlc.arg(names)::text[]),
    unnest(sqlc.arg(spec_values)::text[]);

-- name: CreateMaterialWithSpecs :one
WITH created AS (
    INSERT INTO materials (
        name,
        type
    )
    VALUES (
        $1,
        $2
    )
    RETURNING
        id,
        name,
        type
),
spec AS (
    INSERT INTO material_specs (
        material_id,
        name,
        value
    )
    SELECT
        created.id,
        unnest(sqlc.arg(names)::text[]),
        unnest(sqlc.arg(spec_values)::text[])
    FROM created
)
SELECT
    id,
    name,
    type
FROM created;

-- name: GetMaterialSpecs :many
SELECT
    name,
    value
FROM material_specs
WHERE material_id = $1
ORDER BY name;
