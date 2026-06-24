package dto

type UpdateSKURequest struct {
	Name string `json:"name" binding:"required"`
	Unit string `json:"unit"`
}
