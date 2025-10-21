-- +goose Up
CREATE TABLE texts(
    id UUID PRIMARY KEY,
    value TEXT NOT NULL,
    length INT NOT NULL,
    is_palindrome BOOLEAN NOT NULL,
    word_count INT NOT NULL,
    sha256_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE texts;
