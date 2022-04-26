-- name: CreateUser :one
INSERT INTO "user" (
  role_id, email, telegram_username
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: CreateInitiative :one
INSERT INTO initiative  (
  user_id, question
) VALUES (
  $1, $2
)
RETURNING *;