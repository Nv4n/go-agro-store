// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package db

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type CategoryType string

const (
	CategoryTypePlant CategoryType = "plant"
	CategoryTypeTool  CategoryType = "tool"
	CategoryTypeSeed  CategoryType = "seed"
	CategoryTypeSoil  CategoryType = "soil"
)

func (e *CategoryType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = CategoryType(s)
	case string:
		*e = CategoryType(s)
	default:
		return fmt.Errorf("unsupported scan type for CategoryType: %T", src)
	}
	return nil
}

type NullCategoryType struct {
	CategoryType CategoryType
	Valid        bool // Valid is true if CategoryType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullCategoryType) Scan(value interface{}) error {
	if value == nil {
		ns.CategoryType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.CategoryType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullCategoryType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.CategoryType), nil
}

type ChatStatus string

const (
	ChatStatusOpen   ChatStatus = "open"
	ChatStatusClosed ChatStatus = "closed"
)

func (e *ChatStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ChatStatus(s)
	case string:
		*e = ChatStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for ChatStatus: %T", src)
	}
	return nil
}

type NullChatStatus struct {
	ChatStatus ChatStatus
	Valid      bool // Valid is true if ChatStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullChatStatus) Scan(value interface{}) error {
	if value == nil {
		ns.ChatStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ChatStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullChatStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ChatStatus), nil
}

type DeliveryStatus string

const (
	DeliveryStatusShipped   DeliveryStatus = "shipped"
	DeliveryStatusIntransit DeliveryStatus = "in transit"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusReturned  DeliveryStatus = "returned"
)

func (e *DeliveryStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = DeliveryStatus(s)
	case string:
		*e = DeliveryStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for DeliveryStatus: %T", src)
	}
	return nil
}

type NullDeliveryStatus struct {
	DeliveryStatus DeliveryStatus
	Valid          bool // Valid is true if DeliveryStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullDeliveryStatus) Scan(value interface{}) error {
	if value == nil {
		ns.DeliveryStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.DeliveryStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullDeliveryStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.DeliveryStatus), nil
}

type OrderType string

const (
	OrderTypePending   OrderType = "pending"
	OrderTypeCompleted OrderType = "completed"
	OrderTypeReturned  OrderType = "returned"
)

func (e *OrderType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = OrderType(s)
	case string:
		*e = OrderType(s)
	default:
		return fmt.Errorf("unsupported scan type for OrderType: %T", src)
	}
	return nil
}

type NullOrderType struct {
	OrderType OrderType
	Valid     bool // Valid is true if OrderType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullOrderType) Scan(value interface{}) error {
	if value == nil {
		ns.OrderType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.OrderType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullOrderType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.OrderType), nil
}

type ProdInteractionType string

const (
	ProdInteractionTypeReview   ProdInteractionType = "review"
	ProdInteractionTypeQuestion ProdInteractionType = "question"
)

func (e *ProdInteractionType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ProdInteractionType(s)
	case string:
		*e = ProdInteractionType(s)
	default:
		return fmt.Errorf("unsupported scan type for ProdInteractionType: %T", src)
	}
	return nil
}

type NullProdInteractionType struct {
	ProdInteractionType ProdInteractionType
	Valid               bool // Valid is true if ProdInteractionType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullProdInteractionType) Scan(value interface{}) error {
	if value == nil {
		ns.ProdInteractionType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ProdInteractionType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullProdInteractionType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ProdInteractionType), nil
}

type UserRole string

const (
	UserRoleUser    UserRole = "user"
	UserRoleSupport UserRole = "support"
	UserRoleAdmin   UserRole = "admin"
)

func (e *UserRole) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UserRole(s)
	case string:
		*e = UserRole(s)
	default:
		return fmt.Errorf("unsupported scan type for UserRole: %T", src)
	}
	return nil
}

type NullUserRole struct {
	UserRole UserRole
	Valid    bool // Valid is true if UserRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUserRole) Scan(value interface{}) error {
	if value == nil {
		ns.UserRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UserRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUserRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UserRole), nil
}

type Chat struct {
	ID        pgtype.UUID
	Status    ChatStatus
	CreatedBy pgtype.UUID
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type Delivery struct {
	ID                pgtype.UUID
	OrderID           pgtype.UUID
	Status            DeliveryStatus
	TrackingNumber    pgtype.Text
	EstimatedDelivery pgtype.Timestamptz
	DeliveredAt       pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
}

type Favorite struct {
	UserID    pgtype.UUID
	ProductID pgtype.UUID
	CreatedAt pgtype.Timestamptz
}

type Message struct {
	ID        pgtype.UUID
	ChatID    pgtype.UUID
	UserID    pgtype.UUID
	Content   string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type Order struct {
	ID        pgtype.UUID
	UserID    pgtype.UUID
	Status    OrderType
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type OrderDetail struct {
	ID              pgtype.UUID
	OrderID         pgtype.UUID
	Address         string
	PhoneNumber     pgtype.Text
	ReturnStatement pgtype.Text
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}

type OrderItem struct {
	ID              pgtype.UUID
	OrderID         pgtype.UUID
	ProductID       pgtype.UUID
	Quantity        int32
	PriceAtPurchase pgtype.Numeric
	CreatedAt       pgtype.Timestamptz
}

type Product struct {
	ID          pgtype.UUID
	Name        string
	Price       pgtype.Numeric
	Discount    pgtype.Numeric
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
}

type ProductInteraction struct {
	ID         pgtype.UUID
	ProductID  pgtype.UUID
	UserID     pgtype.UUID
	Type       ProdInteractionType
	Content    string
	IsAnswered pgtype.Bool
	Response   pgtype.Text
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
}

type ProductTag struct {
	ProductID pgtype.UUID
	TagID     pgtype.UUID
}

type Tag struct {
	ID        pgtype.UUID
	Name      string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type User struct {
	ID        pgtype.UUID
	Email     string
	Fname     string
	Lname     string
	Password  string
	Role      UserRole
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}
