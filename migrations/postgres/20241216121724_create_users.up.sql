CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255)
);

CREATE UNIQUE INDEX idx_users_id ON users (id);