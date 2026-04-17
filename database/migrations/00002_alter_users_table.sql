-- +goose up
-- See: https://sqlite.org/lang_altertable.html#making_other_kinds_of_table_schema_changes
CREATE TABLE IF NOT EXISTS new_users (
    id INTEGER PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
) STRICT;
INSERT INTO new_users (id, email, username, password) SELECT id, '', username, password FROM users;
DROP TABLE users;
ALTER TABLE new_users RENAME TO users;

-- +goose down
ALTER TABLE users DROP COLUMN email;
