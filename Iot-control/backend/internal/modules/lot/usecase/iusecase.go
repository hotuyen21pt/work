package usecase

import (
	"context"

	"lot-control/internal/models"
	"lot-control/internal/modules/lot/dto"
)

type IUseCase interface {
	ListBySKU(ctx context.Context, skuID int64) ([]models.Lot, error)
	Upsert(ctx context.Context, params *dto.UpsertLotRequest) (*models.Lot, error)
	Update(ctx context.Context, id int64, params *dto.UpdateLotRequest) (*models.Lot, error)
	Delete(ctx context.Context, id int64) error
}
