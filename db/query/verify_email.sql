-- name: CreateVerifyEmail :one
INSERT INTO verify_emails (
  username,
  email,
  secret_code
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetVerifyEmail :one
SELECT * FROM verify_emails
WHERE id = $1 LIMIT 1;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails
  set 
  is_used = $2
WHERE id = $1
RETURNING *;
