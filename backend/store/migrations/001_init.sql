-- +goose Up
-- The composite PRIMARY KEY also serves single-column user_id lookups, so no
-- separate idx_visits_user_id index is needed.
CREATE TABLE visits (
    user_id      INTEGER  NOT NULL,
    country_name TEXT     NOT NULL,
    added_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, country_name)
);

-- +goose Down
DROP TABLE visits;
