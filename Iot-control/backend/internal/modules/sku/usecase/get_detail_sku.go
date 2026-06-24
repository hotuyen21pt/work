package usecase

import (
	"context"

	"lot-control/internal/models"
)

func (uc *skuUseCase) GetDetail(ctx context.Context, id int64) (*models.SKU, error) {
	result, err := uc.skuRepo.GetDetail(id)
	if err != nil {
		uc.logger.Errorf("GetDetail SKU err: %v", err)
		return nil, err
	}
	return result, nil
}
