// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createChat = `-- name: CreateChat :one
INSERT INTO chats (status, created_by)
VALUES ('open', $1)
RETURNING id
`

func (q *Queries) CreateChat(ctx context.Context, createdBy pgtype.UUID) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createChat, createdBy)
	var id pgtype.UUID
	err := row.Scan(&id)
	return id, err
}

const createMessage = `-- name: CreateMessage :one
INSERT INTO messages (chat_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING id, chat_id, user_id, content, created_at, updated_at
`

type CreateMessageParams struct {
	ChatID  pgtype.UUID
	UserID  pgtype.UUID
	Content string
}

func (q *Queries) CreateMessage(ctx context.Context, arg CreateMessageParams) (Message, error) {
	row := q.db.QueryRow(ctx, createMessage, arg.ChatID, arg.UserID, arg.Content)
	var i Message
	err := row.Scan(
		&i.ID,
		&i.ChatID,
		&i.UserID,
		&i.Content,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createOrder = `-- name: CreateOrder :exec
INSERT INTO orders (user_id, status)
VALUES ($1, $2)
`

type CreateOrderParams struct {
	UserID pgtype.UUID
	Status OrderType
}

func (q *Queries) CreateOrder(ctx context.Context, arg CreateOrderParams) error {
	_, err := q.db.Exec(ctx, createOrder, arg.UserID, arg.Status)
	return err
}

const createProduct = `-- name: CreateProduct :exec
INSERT INTO products (name, price, discount, description, type, category, img)
VALUES ($1, $2, 0, $3, $4, $5, $6)
`

type CreateProductParams struct {
	Name        string
	Price       pgtype.Numeric
	Description pgtype.Text
	Type        pgtype.UUID
	Category    pgtype.UUID
	Img         string
}

func (q *Queries) CreateProduct(ctx context.Context, arg CreateProductParams) error {
	_, err := q.db.Exec(ctx, createProduct,
		arg.Name,
		arg.Price,
		arg.Description,
		arg.Type,
		arg.Category,
		arg.Img,
	)
	return err
}

const createTag = `-- name: CreateTag :one
INSERT INTO tags (name)
VALUES ($1)
RETURNING id, name, created_at, updated_at
`

func (q *Queries) CreateTag(ctx context.Context, name string) (Tag, error) {
	row := q.db.QueryRow(ctx, createTag, name)
	var i Tag
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (email, fname, lname, password, role)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
`

type CreateUserParams struct {
	Email    string
	Fname    string
	Lname    string
	Password string
	Role     UserRole
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Email,
		arg.Fname,
		arg.Lname,
		arg.Password,
		arg.Role,
	)
	var id pgtype.UUID
	err := row.Scan(&id)
	return id, err
}

const deleteChat = `-- name: DeleteChat :exec
DELETE
FROM chats
WHERE id = $1
`

func (q *Queries) DeleteChat(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteChat, id)
	return err
}

const deleteOrder = `-- name: DeleteOrder :exec
DELETE
FROM orders
WHERE id = $1
`

func (q *Queries) DeleteOrder(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteOrder, id)
	return err
}

const deleteOrderItem = `-- name: DeleteOrderItem :exec
DELETE
FROM order_items
WHERE id = $1
`

func (q *Queries) DeleteOrderItem(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteOrderItem, id)
	return err
}

const deleteProduct = `-- name: DeleteProduct :exec
DELETE
FROM products
WHERE id = $1
`

func (q *Queries) DeleteProduct(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteProduct, id)
	return err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1
`

func (q *Queries) DeleteUser(ctx context.Context, id pgtype.UUID) error {
	_, err := q.db.Exec(ctx, deleteUser, id)
	return err
}

const getChatById = `-- name: GetChatById :one
SELECT id, status, created_by, created_at, updated_at
from chats
WHERE id = $1
`

func (q *Queries) GetChatById(ctx context.Context, id pgtype.UUID) (Chat, error) {
	row := q.db.QueryRow(ctx, getChatById, id)
	var i Chat
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CreatedBy,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOpenChatByUserId = `-- name: GetOpenChatByUserId :one
SELECT DISTINCT C.id, C.status, C.created_at, C.updated_at
FROM chats C
         JOIN messages M ON M.chat_id = C.id
WHERE C.status = 'open'
  AND M.user_id = $1
`

type GetOpenChatByUserIdRow struct {
	ID        pgtype.UUID
	Status    ChatStatus
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (q *Queries) GetOpenChatByUserId(ctx context.Context, userID pgtype.UUID) (GetOpenChatByUserIdRow, error) {
	row := q.db.QueryRow(ctx, getOpenChatByUserId, userID)
	var i GetOpenChatByUserIdRow
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOrderById = `-- name: GetOrderById :one
SELECT id, user_id, status, created_at, updated_at
FROM orders
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetOrderById(ctx context.Context, id pgtype.UUID) (Order, error) {
	row := q.db.QueryRow(ctx, getOrderById, id)
	var i Order
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOrderDetailsById = `-- name: GetOrderDetailsById :one
SELECT id, order_id, address, phone_number, return_statement, created_at, updated_at
FROM order_details
WHERE order_id = $1
LIMIT 1
`

func (q *Queries) GetOrderDetailsById(ctx context.Context, orderID pgtype.UUID) (OrderDetail, error) {
	row := q.db.QueryRow(ctx, getOrderDetailsById, orderID)
	var i OrderDetail
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Address,
		&i.PhoneNumber,
		&i.ReturnStatement,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getOrderItemById = `-- name: GetOrderItemById :one
SELECT id, order_id, address, phone_number, return_statement, created_at, updated_at
FROM order_details
WHERE id = $1
`

func (q *Queries) GetOrderItemById(ctx context.Context, id pgtype.UUID) (OrderDetail, error) {
	row := q.db.QueryRow(ctx, getOrderItemById, id)
	var i OrderDetail
	err := row.Scan(
		&i.ID,
		&i.OrderID,
		&i.Address,
		&i.PhoneNumber,
		&i.ReturnStatement,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getProductById = `-- name: GetProductById :one
SELECT DISTINCT P.id,
                P.name,
                P.price,
                P.discount,
                P.description,
                P.created_at,
                P.updated_at,
                P.img,
                TYP.name as type,
                CAT.name as category
FROM products P
         JOIN tags TYP on TYP.id = P.type
         JOIN tags CAT on CAT.id = P.category
WHERE P.id = $1
LIMIT 1
`

type GetProductByIdRow struct {
	ID          pgtype.UUID
	Name        string
	Price       pgtype.Numeric
	Discount    pgtype.Numeric
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	Img         string
	Type        string
	Category    string
}

func (q *Queries) GetProductById(ctx context.Context, id pgtype.UUID) (GetProductByIdRow, error) {
	row := q.db.QueryRow(ctx, getProductById, id)
	var i GetProductByIdRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.Discount,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Img,
		&i.Type,
		&i.Category,
	)
	return i, err
}

const getProductByName = `-- name: GetProductByName :one
SELECT DISTINCT P.id,
                P.name,
                P.price,
                P.discount,
                P.description,
                P.created_at,
                P.updated_at,
                P.img,
                TYP.name as type,
                CAT.name as category
FROM products P
         JOIN tags TYP on TYP.id = P.type
         JOIN tags CAT on CAT.id = P.category
WHERE P.name = $1
LIMIT 1
`

type GetProductByNameRow struct {
	ID          pgtype.UUID
	Name        string
	Price       pgtype.Numeric
	Discount    pgtype.Numeric
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	Img         string
	Type        string
	Category    string
}

func (q *Queries) GetProductByName(ctx context.Context, name string) (GetProductByNameRow, error) {
	row := q.db.QueryRow(ctx, getProductByName, name)
	var i GetProductByNameRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.Discount,
		&i.Description,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Img,
		&i.Type,
		&i.Category,
	)
	return i, err
}

const getTagById = `-- name: GetTagById :one
SELECT id, name, created_at, updated_at
FROM tags
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetTagById(ctx context.Context, id pgtype.UUID) (Tag, error) {
	row := q.db.QueryRow(ctx, getTagById, id)
	var i Tag
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getTagByName = `-- name: GetTagByName :one
SELECT id, name, created_at, updated_at
FROM tags
WHERE name = $1
LIMIT 1
`

func (q *Queries) GetTagByName(ctx context.Context, name string) (Tag, error) {
	row := q.db.QueryRow(ctx, getTagByName, name)
	var i Tag
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, email, password
FROM users
WHERE email = $1
LIMIT 1
`

type GetUserByEmailRow struct {
	ID       pgtype.UUID
	Email    string
	Password string
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i GetUserByEmailRow
	err := row.Scan(&i.ID, &i.Email, &i.Password)
	return i, err
}

const getUserById = `-- name: GetUserById :one
SELECT id, email, fname, lname, role
FROM users
WHERE id = $1
LIMIT 1
`

type GetUserByIdRow struct {
	ID    pgtype.UUID
	Email string
	Fname string
	Lname string
	Role  UserRole
}

func (q *Queries) GetUserById(ctx context.Context, id pgtype.UUID) (GetUserByIdRow, error) {
	row := q.db.QueryRow(ctx, getUserById, id)
	var i GetUserByIdRow
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Fname,
		&i.Lname,
		&i.Role,
	)
	return i, err
}

const listAllCategoryTags = `-- name: ListAllCategoryTags :many
SELECT DISTINCT P.id, P.name
FROM tags T
         JOIN products P ON T.id = P.category
`

type ListAllCategoryTagsRow struct {
	ID   pgtype.UUID
	Name string
}

func (q *Queries) ListAllCategoryTags(ctx context.Context) ([]ListAllCategoryTagsRow, error) {
	rows, err := q.db.Query(ctx, listAllCategoryTags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAllCategoryTagsRow
	for rows.Next() {
		var i ListAllCategoryTagsRow
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllChats = `-- name: ListAllChats :one
SELECT id, status, created_by, created_at, updated_at
FROM chats
`

func (q *Queries) ListAllChats(ctx context.Context) (Chat, error) {
	row := q.db.QueryRow(ctx, listAllChats)
	var i Chat
	err := row.Scan(
		&i.ID,
		&i.Status,
		&i.CreatedBy,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listAllMessagesByChatId = `-- name: ListAllMessagesByChatId :one
SELECT id, chat_id, user_id, content, created_at, updated_at
FROM messages
WHERE chat_id = $1
ORDER BY created_at desc
`

func (q *Queries) ListAllMessagesByChatId(ctx context.Context, chatID pgtype.UUID) (Message, error) {
	row := q.db.QueryRow(ctx, listAllMessagesByChatId, chatID)
	var i Message
	err := row.Scan(
		&i.ID,
		&i.ChatID,
		&i.UserID,
		&i.Content,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listAllOrderItemsById = `-- name: ListAllOrderItemsById :many
SELECT id, order_id, product_id, quantity, price_at_purchase, created_at
FROM order_items
WHERE order_id = $1
`

func (q *Queries) ListAllOrderItemsById(ctx context.Context, orderID pgtype.UUID) ([]OrderItem, error) {
	rows, err := q.db.Query(ctx, listAllOrderItemsById, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OrderItem
	for rows.Next() {
		var i OrderItem
		if err := rows.Scan(
			&i.ID,
			&i.OrderID,
			&i.ProductID,
			&i.Quantity,
			&i.PriceAtPurchase,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllOrders = `-- name: ListAllOrders :many
SELECT id, user_id, status, created_at, updated_at
FROM orders
`

func (q *Queries) ListAllOrders(ctx context.Context) ([]Order, error) {
	rows, err := q.db.Query(ctx, listAllOrders)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Order
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllOrdersByUserId = `-- name: ListAllOrdersByUserId :many
SELECT id, user_id, status, created_at, updated_at
FROM orders
WHERE user_id = $1
`

func (q *Queries) ListAllOrdersByUserId(ctx context.Context, userID pgtype.UUID) ([]Order, error) {
	rows, err := q.db.Query(ctx, listAllOrdersByUserId, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Order
	for rows.Next() {
		var i Order
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllProducts = `-- name: ListAllProducts :many
SELECT DISTINCT P.id,
                P.name,
                P.price,
                P.discount,
                P.description,
                P.created_at,
                P.updated_at,
                P.img,
                TYP.name as type,
                CAT.name as category
FROM products P
         JOIN tags TYP on TYP.id = P.type
         JOIN tags CAT on CAT.id = P.category
`

type ListAllProductsRow struct {
	ID          pgtype.UUID
	Name        string
	Price       pgtype.Numeric
	Discount    pgtype.Numeric
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	Img         string
	Type        string
	Category    string
}

func (q *Queries) ListAllProducts(ctx context.Context) ([]ListAllProductsRow, error) {
	rows, err := q.db.Query(ctx, listAllProducts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAllProductsRow
	for rows.Next() {
		var i ListAllProductsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Price,
			&i.Discount,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Img,
			&i.Type,
			&i.Category,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllProductsByType = `-- name: ListAllProductsByType :many
SELECT DISTINCT P.id,
                P.name,
                P.price,
                P.discount,
                P.description,
                P.created_at,
                P.updated_at,
                P.img,
                TYP.name as type,
                CAT.name as category
FROM products P
         JOIN tags TYP on TYP.id = P.type
         JOIN tags CAT on CAT.id = P.category
WHERE TYP.name = $1
`

type ListAllProductsByTypeRow struct {
	ID          pgtype.UUID
	Name        string
	Price       pgtype.Numeric
	Discount    pgtype.Numeric
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	Img         string
	Type        string
	Category    string
}

func (q *Queries) ListAllProductsByType(ctx context.Context, name string) ([]ListAllProductsByTypeRow, error) {
	rows, err := q.db.Query(ctx, listAllProductsByType, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAllProductsByTypeRow
	for rows.Next() {
		var i ListAllProductsByTypeRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Price,
			&i.Discount,
			&i.Description,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Img,
			&i.Type,
			&i.Category,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllTags = `-- name: ListAllTags :many
SELECT id, name, created_at, updated_at
FROM tags
`

func (q *Queries) ListAllTags(ctx context.Context) ([]Tag, error) {
	rows, err := q.db.Query(ctx, listAllTags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Tag
	for rows.Next() {
		var i Tag
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAllUsers = `-- name: ListAllUsers :many
SELECT id, email, fname, lname, role
FROM users
ORDER BY fname, lname
`

type ListAllUsersRow struct {
	ID    pgtype.UUID
	Email string
	Fname string
	Lname string
	Role  UserRole
}

func (q *Queries) ListAllUsers(ctx context.Context) ([]ListAllUsersRow, error) {
	rows, err := q.db.Query(ctx, listAllUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAllUsersRow
	for rows.Next() {
		var i ListAllUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Email,
			&i.Fname,
			&i.Lname,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateChatStatus = `-- name: UpdateChatStatus :exec
UPDATE chats
SET status = $2
WHERE id = $1
`

type UpdateChatStatusParams struct {
	ID     pgtype.UUID
	Status ChatStatus
}

func (q *Queries) UpdateChatStatus(ctx context.Context, arg UpdateChatStatusParams) error {
	_, err := q.db.Exec(ctx, updateChatStatus, arg.ID, arg.Status)
	return err
}

const updateOrderDetails = `-- name: UpdateOrderDetails :exec
UPDATE order_details
SET address         = $2,
    phone_number=$3,
    return_statement=$4
WHERE order_id = $1
`

type UpdateOrderDetailsParams struct {
	OrderID         pgtype.UUID
	Address         string
	PhoneNumber     pgtype.Text
	ReturnStatement pgtype.Text
}

func (q *Queries) UpdateOrderDetails(ctx context.Context, arg UpdateOrderDetailsParams) error {
	_, err := q.db.Exec(ctx, updateOrderDetails,
		arg.OrderID,
		arg.Address,
		arg.PhoneNumber,
		arg.ReturnStatement,
	)
	return err
}

const updateOrderItemById = `-- name: UpdateOrderItemById :exec
UPDATE order_items
SET quantity=$2,
    price_at_purchase=$3
WHERE id = $1
`

type UpdateOrderItemByIdParams struct {
	ID              pgtype.UUID
	Quantity        int32
	PriceAtPurchase pgtype.Numeric
}

func (q *Queries) UpdateOrderItemById(ctx context.Context, arg UpdateOrderItemByIdParams) error {
	_, err := q.db.Exec(ctx, updateOrderItemById, arg.ID, arg.Quantity, arg.PriceAtPurchase)
	return err
}

const updateOrderStatus = `-- name: UpdateOrderStatus :exec
UPDATE orders
SET status =$2
WHERE id = $1
`

type UpdateOrderStatusParams struct {
	ID     pgtype.UUID
	Status OrderType
}

func (q *Queries) UpdateOrderStatus(ctx context.Context, arg UpdateOrderStatusParams) error {
	_, err := q.db.Exec(ctx, updateOrderStatus, arg.ID, arg.Status)
	return err
}

const updateProduct = `-- name: UpdateProduct :exec
UPDATE products
SET name= $2,
    price=$3,
    discount=$4,
    description=$5,
    img=$6,
    category=$7,
    type=$8
WHERE id = $1
`

type UpdateProductParams struct {
	ID          pgtype.UUID
	Name        string
	Price       pgtype.Numeric
	Discount    pgtype.Numeric
	Description pgtype.Text
	Img         string
	Category    pgtype.UUID
	Type        pgtype.UUID
}

func (q *Queries) UpdateProduct(ctx context.Context, arg UpdateProductParams) error {
	_, err := q.db.Exec(ctx, updateProduct,
		arg.ID,
		arg.Name,
		arg.Price,
		arg.Discount,
		arg.Description,
		arg.Img,
		arg.Category,
		arg.Type,
	)
	return err
}

const updateUserNames = `-- name: UpdateUserNames :one
UPDATE users
SET fname = $2,
    lname = $3
WHERE id = $1
RETURNING id, email, fname, lname, password, role, created_at, updated_at
`

type UpdateUserNamesParams struct {
	ID    pgtype.UUID
	Fname string
	Lname string
}

func (q *Queries) UpdateUserNames(ctx context.Context, arg UpdateUserNamesParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserNames, arg.ID, arg.Fname, arg.Lname)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Fname,
		&i.Lname,
		&i.Password,
		&i.Role,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserPass = `-- name: UpdateUserPass :one
UPDATE users
SET password = $2
WHERE id = $1
RETURNING id, email, fname, lname, password, role, created_at, updated_at
`

type UpdateUserPassParams struct {
	ID       pgtype.UUID
	Password string
}

func (q *Queries) UpdateUserPass(ctx context.Context, arg UpdateUserPassParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserPass, arg.ID, arg.Password)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Fname,
		&i.Lname,
		&i.Password,
		&i.Role,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUserRole = `-- name: UpdateUserRole :one
UPDATE users
SET role = $2
WHERE id = $1
RETURNING id, email, fname, lname, password, role, created_at, updated_at
`

type UpdateUserRoleParams struct {
	ID   pgtype.UUID
	Role UserRole
}

func (q *Queries) UpdateUserRole(ctx context.Context, arg UpdateUserRoleParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserRole, arg.ID, arg.Role)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.Fname,
		&i.Lname,
		&i.Password,
		&i.Role,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
