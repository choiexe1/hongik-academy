-- +goose Up
-- +goose StatementBegin
CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    gender VARCHAR(1) NOT NULL CHECK (gender IN ('M', 'F')),
    phone VARCHAR(20),
    parent_phone VARCHAR(20),
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_students_name ON students(name);
CREATE INDEX idx_students_gender ON students(gender);
CREATE INDEX idx_students_phone ON students(phone);
CREATE INDEX idx_students_parent_phone ON students(parent_phone);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS students;
-- +goose StatementEnd
