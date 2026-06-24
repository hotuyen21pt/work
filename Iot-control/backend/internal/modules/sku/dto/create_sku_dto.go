package dto

type CreateSKURequest struct {
	SKUCode string `json:"sku_code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Unit    string `json:"unit"`
}
