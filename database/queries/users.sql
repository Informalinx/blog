-- name: FindByUsername :one
SELECT id, username, password FROM users WHERE users.username = ?;

-- name: CreateUser :exec
INSERT INTO users (
    username,
    password
) VALUES (
    ?, ?
);
