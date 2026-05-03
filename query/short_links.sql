-- name: CreateShortLink :one
INSERT INTO
    short_links (original_url, short_name, short_url)
VALUES
    (
        sqlc.arg (original_url),
        sqlc.arg (short_name),
        sqlc.arg (short_url)
    ) RETURNING id,
    original_url,
    short_name,
    short_url,
    created_at;

-- name: GetShortLinkById :one
SELECT
    id,
    original_url,
    short_name,
    short_url,
    created_at
FROM
    short_links
WHERE
    id = sqlc.arg (id);

-- name: GetShortLinks :many
SELECT
    id,
    original_url,
    short_name,
    short_url,
    created_at
FROM
    short_links;

-- name: UpdateShortLink :one
UPDATE short_links
SET
    original_url = sqlc.arg (original_url),
    short_name = sqlc.arg (short_name),
    short_url = sqlc.arg (short_url)
WHERE
    id = sqlc.arg (id) RETURNING id,
    original_url,
    short_name,
    short_url,
    created_at;

-- name: DeleteShortLink :one
DELETE FROM short_links
WHERE
    id = sqlc.arg (id) RETURNING id,
    original_url,
    short_name,
    short_url,
    created_at;
