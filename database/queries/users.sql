-- name: FindByEmail :one
SELECT id, username, email, password FROM users WHERE users.email = ?;

-- name: CreateUser :exec
INSERT INTO users (
    username,
    email,
    password
) VALUES (
    ?, ?, ?
);
