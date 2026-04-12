-- +goose up
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    password TEXT NOT NULL
) STRICT;

-- +goose down
DROP TABLE users;
