-- +goose Up
ALTER TABLE link_visits
ADD COLUMN id SERIAL;

ALTER TABLE link_visits ADD PRIMARY KEY (id);

-- +goose Down
ALTER TABLE link_visits
DROP COLUMN IF EXISTS id;
