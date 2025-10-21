-- name: CreateText :one
INSERT INTO texts (id, value, length, is_palindrome, word_count, sha256_hash, created_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    $5,
    NOW()
)
RETURNING id;

-- name: CreateCharCount :exec
INSERT INTO character_count (id, string_id, character, char_count)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3
);

-- name: GetText :one
SELECT value 
FROM texts WHERE value = $1;

-- name: DeleteText :exec
DELETE FROM texts
WHERE id = $1;




