-- +goose Up
CREATE TABLE visits (
    user_id      INTEGER  NOT NULL,
    country_name TEXT     NOT NULL,
    added_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, country_name)
);
CREATE INDEX idx_visits_user_id ON visits(user_id);

-- +goose Down
DROP INDEX idx_visits_user_id;
DROP TABLE visits;
