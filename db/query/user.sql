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
    hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
    firstname = COALESCE(sqlc.narg(firstname), firstname),    
    lastname = COALESCE(sqlc.narg(lastname), lastname),    
    email = COALESCE(sqlc.narg(email), email),
    password_changed_at = COALESCE(sqlc.narg(password_changed_at), password_changed_at)
WHERE username = sqlc.arg(username)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE username = $1;
