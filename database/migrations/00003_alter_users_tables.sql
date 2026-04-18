-- +goose up
-- See: https://sqlite.org/lang_altertable.html#making_other_kinds_of_table_schema_changes
CREATE TABLE IF NOT EXISTS new_users (
    id INTEGER PRIMARY KEY,
    -- encrypted email : can be decrypted using a key to send recover password emails
    email TEXT NOT NULL UNIQUE,
    -- hashed email : can be used to make a lookup and check if a user with a particular email exists.
    -- hash must be replicable : hash of "email@test.com" must always return the same value for example
    email_hash TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
) STRICT;
INSERT INTO new_users (id, email, email_hash, username, password) SELECT id, email, '', username, password FROM users;
DROP TABLE users;
ALTER TABLE new_users RENAME TO users;

-- +goose down
ALTER TABLE users DROP COLUMN email_hash;
