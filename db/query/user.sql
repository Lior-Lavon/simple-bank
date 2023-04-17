-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  firstname, 
  lastname, 
  email
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY username
LIMIT $1
OFFSET $2;

-- name: UpdateUser :one
UPDATE users
  set 
  firstname = $2, 
  lastname = $3, 
  email = $4
WHERE username = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE username = $1;
