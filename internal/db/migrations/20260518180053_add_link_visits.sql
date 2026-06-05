-- +goose Up
CREATE TABLE
    IF NOT EXISTS link_visits (
        ip inet NOT NULL,
        user_agent text NOT NULL,
        referer text NOT NULL,
        status integer NOT NULL,
        created_at timestamptz NOT NULL DEFAULT now ()
    );

-- +goose Down
DROP TABLE IF EXISTS link_visits;
