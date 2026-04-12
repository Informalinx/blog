-- name: CreateUser :exec
INSERT INTO users (
    username,
    password
) VALUES (
    ?, ?
);
