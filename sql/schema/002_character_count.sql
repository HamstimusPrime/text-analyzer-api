-- +goose Up
CREATE TABLE character_count(
    id UUID PRIMARY KEY,
    string_id UUID NOT NULL,
    character TEXT NOT NULL UNIQUE,
    char_count INTEGER NOT NULL,
    CONSTRAINT fk_character_id
        FOREIGN KEY(string_id)
        REFERENCES texts(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP TABLE character_count;