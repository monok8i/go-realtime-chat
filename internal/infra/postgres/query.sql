-- name: CreateMessage :one
INSERT INTO messages (user_id, chat_id, text)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMessagesByChat :many
SELECT * FROM messages
WHERE chat_id = $1
ORDER BY id
LIMIT $2 OFFSET $3;
