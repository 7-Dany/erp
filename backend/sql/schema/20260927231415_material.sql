-- +goose Up
CREATE TYPE MATERIAL_TYPE AS ENUM('unspecified', 'raw_material', 'packaging', 'semi_finished', 'finished_product', 'spare_part');

CREATE TABLE materials(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type MATERIAL_TYPE NOT NULL
);

CREATE INDEX materials_name_trgm_idx ON materials USING gin (name gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS materials_name_trgm_idx;
DROP TABLE materials;
DROP TYPE MATERIAL_TYPE;
