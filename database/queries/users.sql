-- name: FindByEmail :one
SELECT id, username, email, email_hash, password FROM users WHERE users.email_hash = ?;

-- name: CreateUser :exec
INSERT INTO users (
    username,
    email,
    email_hash,
    password
) VALUES (
    ?, ?, ?, ?
);
