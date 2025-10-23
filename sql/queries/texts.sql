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
INSERT INTO character_count (id, string_id, character, unique_char_count)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3
);

-- name: GetText :one
SELECT id, value, length, is_palindrome, word_count, sha256_hash, created_at 
FROM texts WHERE value = $1;

-- name: GetTextByID :one
SELECT id, value, length, is_palindrome, word_count, sha256_hash, created_at 
FROM texts WHERE id = $1;

-- name: GetAllTexts :many
SELECT id, value, length, is_palindrome, word_count, sha256_hash, created_at 
FROM texts 
ORDER BY created_at DESC;

-- name: GetCharacterCountsByID :many
SELECT character, unique_char_count
FROM texts
JOIN character_count ON $1 = character_count.string_id
ORDER BY character;


-- name: GetFilteredTexts :many
SELECT DISTINCT
    t.id,
    t.value,
    t.length,
    t.is_palindrome,
    t.word_count,
    t.sha256_hash,
    t.created_at
FROM texts t
WHERE 
    t.is_palindrome = @is_palindrome AND
    t.length >= @min_length AND
    t.length <= @max_length AND
    t.word_count = @word_count AND
    t.value LIKE '%' || @contains_character || '%'
ORDER BY t.created_at DESC;


-- name: DeleteTextWithValue :exec
DELETE FROM texts
WHERE value = $1;

-- name: DeleteTextWithID :exec
DELETE FROM texts
WHERE id = $1;






