-- +goose Up

CREATE TABLE IF NOT EXISTS url (
    id SERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down

DROP TABLE IF EXISTS url;