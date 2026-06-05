-- +goose Up
ALTER TABLE link_visits
ADD COLUMN link_id INTEGER;

-- +goose Down
ALTER TABLE link_visits
DROP COLUMN IF EXISTS link_id;
