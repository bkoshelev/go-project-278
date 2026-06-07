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

-- name: GetShortLinkByShortName :one
SELECT
    id,
    original_url,
    short_name,
    short_url,
    created_at
FROM
    short_links
WHERE
    short_name = sqlc.arg (short_name);

-- name: GetShortLinks :many
SELECT
    id,
    original_url,
    short_name,
    short_url,
    created_at
FROM
    short_links
ORDER BY id
LIMIT
    COALESCE(sqlc.narg('limit'), 20)
OFFSET $1;

-- name: UpdateShortLink :one
UPDATE short_links
SET
    original_url = sqlc.arg (original_url),
    short_name = sqlc.arg (short_name),
    short_url = sqlc.arg (short_url)
WHERE
    id = sqlc.arg (id)
RETURNING id,
    original_url,
    short_name,
    short_url;

-- name: DeleteShortLink :one
DELETE FROM short_links
WHERE
    id = sqlc.arg (id) RETURNING id,
    original_url,
    short_name,
    short_url,
    created_at;

-- name: CountShortLinks :one
SELECT count(*) FROM short_links;

-- name: CreateLinkVisit :one
INSERT INTO
    link_visits (ip, link_id, user_agent, referer, status)
VALUES
    (
        sqlc.arg (ip),
        sqlc.arg (link_id),
        sqlc.arg (user_agent),
        sqlc.arg (referer),
        sqlc.arg (status)
    ) RETURNING id,
    ip,
    link_id,
    user_agent,
    referer,
    status,
    created_at;

-- name: GetLinkVisits :many
SELECT
    id,
    ip,
    link_id,
    user_agent,
    referer AS reffer,
    status,
    created_at
FROM
    link_visits
ORDER BY id
LIMIT
    COALESCE(sqlc.narg('limit'), 20)
OFFSET $1;

-- name: CountLinkVisits :one
SELECT count(*) FROM link_visits;
