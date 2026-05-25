-- +goose Up
CREATE TABLE employees (
   id SERIAL PRIMARY KEY,
   department_id INTEGER NOT NULL,
   full_name VARCHAR(255) NOT NULL,
   "position" VARCHAR(255) NOT NULL,
   hired_at DATE,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
   CONSTRAINT fk_employee_department
       FOREIGN KEY(department_id)
           REFERENCES departments(id)
           ON DELETE CASCADE
);

CREATE INDEX idx_employees_department_id ON employees(department_id);

-- +goose Down
DROP TABLE IF EXISTS employees;
