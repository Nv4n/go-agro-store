-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;

-- name: ListAllUsers :many
SELECT id, email, fname, lname
FROM users
ORDER BY fname, lname;

-- name: CreateUser :one
INSERT INTO users (email, fname, lname, password, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: UpdateUserNames :one
UPDATE users
SET fname = $2,
    lname = $3
WHERE id = $1
RETURNING *;

-- name: UpdateUserPass :one
UPDATE users
SET password = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;

-- name: GetChatById :one
SELECT *
from chats
WHERE id = $1;

-- name: GetOpenChatByUserId :one
SELECT DISTINCT C.id, C.status, C.created_at, C.updated_at
FROM chats C
         JOIN messages M ON M.chat_id = C.id
WHERE C.status = 'open'
  AND M.user_id = $1;

-- name: ListAllChats :one
SELECT *
FROM chats;

-- name: CreateChat :one
INSERT INTO chats (status)
VALUES ('open')
RETURNING id;

-- name: UpdateChatStatus :exec
UPDATE chats SET status = $2
WHERE id = $1;

-- name: DeleteChat :one
DELETE FROM chats
WHERE id = $1;

-- name: ListAllMessagesByChatId :one
SELECT *
FROM messages
WHERE chat_id = $1
ORDER BY created_at desc;

-- name: CreateMessage :one
INSERT INTO messages (chat_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListAllOrders :many
SELECT * FROM orders;

-- name: ListAllOrdersByUserId :many
SELECT * FROM orders
WHERE user_id=$1;

-- name: GetOrderById :one
SELECT * FROM orders
WHERE  id=$1 LIMIT 1;

-- name: UpdateOrderStatus :exec
UPDATE orders SET status =$2
WHERE  id = $1;

-- name: DeleteOrder :exec
DELETE from orders
where id=$1;
