package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Branch       string    `json:"branch"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type SKU struct {
	ID        int64     `json:"id"`
	SKUCode   string    `json:"sku_code"`
	Name      string    `json:"name"`
	Unit      string    `json:"unit"`
	TotalQty  int       `json:"total_qty"`
	LotCount  int       `json:"lot_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Lots      []Lot     `json:"lots,omitempty"`
}

type Lot struct {
	ID              int64     `json:"id"`
	SKUID           int64     `json:"sku_id"`
	SKUCode         string    `json:"sku_code,omitempty"`
	LotNumber       string    `json:"lot_number"`
	ManufactureDate string    `json:"manufacture_date"`
	ExpiryDate      string    `json:"expiry_date"`
	Qty             int       `json:"qty"`
	Branch          string    `json:"branch"`
	CountedByID     *int64    `json:"counted_by"`
	CountedByName   string    `json:"counted_by_name"`
	CountedAt       time.Time `json:"counted_at"`
	Notes           string    `json:"notes"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateSKURequest struct {
	SKUCode string `json:"sku_code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Unit    string `json:"unit"`
}

type UpdateSKURequest struct {
	Name string `json:"name" binding:"required"`
	Unit string `json:"unit"`
}

type UpsertLotRequest struct {
	SKUID           int64  `json:"sku_id" binding:"required"`
	LotNumber       string `json:"lot_number" binding:"required"`
	ManufactureDate string `json:"manufacture_date"`
	ExpiryDate      string `json:"expiry_date"`
	Qty             int    `json:"qty" binding:"min=0"`
	Branch          string `json:"branch"`
	Notes           string `json:"notes"`
}

type UpdateLotRequest struct {
	ManufactureDate string `json:"manufacture_date"`
	ExpiryDate      string `json:"expiry_date"`
	Qty             int    `json:"qty" binding:"min=0"`
	Notes           string `json:"notes"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Branch   string `json:"branch"`
	Role     string `json:"role"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name" binding:"required"`
	Branch   string `json:"branch" binding:"required"`
	Role     string `json:"role"`
}
