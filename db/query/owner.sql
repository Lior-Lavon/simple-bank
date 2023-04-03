-- name: CreateOwner :one
INSERT INTO owners (
  firstname, 
  lastname, 
  email
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetOwner :one
SELECT * FROM owners
WHERE id = $1 LIMIT 1;

-- name: ListOwners :many
SELECT * FROM owners
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: UpdateOwner :one
UPDATE owners
  set 
  firstname = $2, 
  lastname = $3, 
  email = $4
WHERE id = $1
RETURNING *;

-- name: DeleteOwner :exec
DELETE FROM owners
WHERE id = $1;
