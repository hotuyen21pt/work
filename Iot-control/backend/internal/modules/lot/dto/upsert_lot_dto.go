package dto

type UpsertLotRequest struct {
	SKUID           int64  `json:"sku_id" binding:"required"`
	LotNumber       string `json:"lot_number" binding:"required"`
	ManufactureDate string `json:"manufacture_date"`
	ExpiryDate      string `json:"expiry_date"`
	Qty             int    `json:"qty" binding:"min=0"`
	Branch          string `json:"branch"`
	Notes           string `json:"notes"`

	// CountedByID được handler nạp từ context (người đang đăng nhập), không bind từ body.
	CountedByID *int64 `json:"-"`
}

type UpdateLotRequest struct {
	ManufactureDate string `json:"manufacture_date"`
	ExpiryDate      string `json:"expiry_date"`
	Qty             int    `json:"qty" binding:"min=0"`
	Notes           string `json:"notes"`

	// CountedByID được handler nạp từ context (người đang đăng nhập), không bind từ body.
	CountedByID *int64 `json:"-"`
}
