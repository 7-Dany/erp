-- +goose Up
CREATE TYPE MATERIAL_TYPE AS ENUM('unspecified', 'raw_material', 'packaging', 'semi_finished', 'finished_product', 'spare_part', 'consumable');

CREATE TABLE materials(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type MATERIAL_TYPE NOT NULL
);

CREATE INDEX materials_name_trgm_idx ON materials USING gin (name gin_trgm_ops);

CREATE TABLE material_specs(
    material_id BIGINT NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    name VARCHAR(64) NOT NULL,
    value VARCHAR(255) NOT NULL,
    PRIMARY KEY (material_id, name)
);

CREATE UNIQUE INDEX material_specs_name_ci_idx ON material_specs (material_id, lower(name));

-- +goose Down
DROP INDEX IF EXISTS material_specs_name_ci_idx;
DROP TABLE IF EXISTS material_specs;
DROP INDEX IF EXISTS materials_name_trgm_idx;
DROP TABLE materials;
DROP TYPE MATERIAL_TYPE;
