package usecase

import (
	"context"

	"lot-control/internal/models"
	"lot-control/internal/modules/sku/dto"
)

type IUseCase interface {
	List(ctx context.Context, q string) ([]models.SKU, error)
	Create(ctx context.Context, params *dto.CreateSKURequest) (*models.SKU, error)
	GetDetail(ctx context.Context, id int64) (*models.SKU, error)
	Update(ctx context.Context, id int64, params *dto.UpdateSKURequest) error
	Delete(ctx context.Context, id int64) error
}
