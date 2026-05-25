-- +goose Up
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    parent_id INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_parent_department
     FOREIGN KEY(parent_id)
         REFERENCES departments(id)
         ON DELETE SET NULL
);

CREATE INDEX idx_departments_parent_id ON departments(parent_id);

-- +goose Down
DROP TABLE IF EXISTS departments;
