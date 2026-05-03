-- +goose Up
CREATE TABLE
    IF NOT EXISTS short_links (
        id serial PRIMARY KEY,
        original_url text NOT NULL,
        short_name text UNIQUE NOT NULL,
        short_url text UNIQUE NOT NULL,
        created_at timestamptz NOT NULL DEFAULT now ()
    );

-- +goose Down
DROP TABLE IF EXISTS short_links;
