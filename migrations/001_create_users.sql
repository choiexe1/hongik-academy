-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 기본 최고관리자 계정 (password: admin123)
INSERT INTO users (username, name, password_hash, role) VALUES
('admin', '최고관리자', '$2a$10$R6622lqkuY9SYIZnBeCVTeTkiCfa.xP6FcKYzt0ChI0rSQ0ScUSzi', 'super_admin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
