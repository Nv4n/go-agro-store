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

-- name: UpdateUserRole :one
UPDATE users
SET role = $2
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
INSERT INTO chats (status, created_by)
VALUES ('open', $1)
RETURNING id;

-- name: UpdateChatStatus :exec
UPDATE chats
SET status = $2
WHERE id = $1;

-- name: DeleteChat :exec
DELETE
FROM chats
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
SELECT *
FROM orders;

-- name: ListAllOrdersByUserId :many
SELECT *
FROM orders
WHERE user_id = $1;

-- name: GetOrderById :one
SELECT *
FROM orders
WHERE id = $1
LIMIT 1;

-- name: CreateOrder :exec
INSERT INTO orders (user_id, status)
VALUES ($1, $2);

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status =$2
WHERE id = $1;

-- name: DeleteOrder :exec
DELETE
FROM orders
WHERE id = $1;

-- name: GetOrderDetailsById :one
SELECT *
FROM order_details
WHERE order_id = $1
LIMIT 1;

-- name: UpdateOrderDetails :exec
UPDATE order_details
SET address         = $2,
    phone_number=$3,
    return_statement=$4
WHERE order_id = $1;

-- name: ListAllOrderItemsById :many
SELECT *
FROM order_items
WHERE order_id = $1;

-- name: GetOrderItemById :one
SELECT *
FROM order_details
WHERE id = $1;

-- name: UpdateOrderItemById :exec
UPDATE order_items
SET quantity=$2,
    price_at_purchase=$3
WHERE id = $1;

-- name: DeleteOrderItem :exec
DELETE
FROM order_items
WHERE id = $1;

-- name: ListAllProducts :many
SELECT DISTINCT P.id,
                P.name,
                P.price,
                P.discount,
                P.description,
                P.created_at,
                P.updated_at,
                TYP.name as type,
                CAT.name as category
FROM products P
         JOIN tags TYP on TYP.id = P.type
         JOIN tags CAT on CAT.id = P.category;

-- name: GetProductById :one
SELECT DISTINCT P.id,
                P.name,
                P.price,
                P.discount,
                P.description,
                P.created_at,
                P.updated_at,
                TYP.name as type,
                CAT.name as category
FROM products P
         JOIN tags TYP on TYP.id = P.type
         JOIN tags CAT on CAT.id = P.category
WHERE P.id = $1
LIMIT 1;

-- name: UpdateProduct :exec
UPDATE products
SET name= $2,
    price=$3,
    discount=$4,
    description=$5
WHERE id = $1;

-- name: DeleteProduct :exec
DELETE
FROM products
WHERE id = $1;