-- +goose Up
CREATE TYPE MATERIAL_TYPE AS ENUM('unspecified', 'raw_material', 'packaging', 'semi_finished', 'finished_product', 'spare_part');

CREATE TABLE materials(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type MATERIAL_TYPE NOT NULL
);

-- +goose Down
DROP TABLE materials;
DROP TYPE MATERIAL_TYPE;
