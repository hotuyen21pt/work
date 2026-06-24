package usecase

import (
	"context"

	"lot-control/internal/models"
	"lot-control/internal/modules/sku/dto"
)

func (uc *skuUseCase) Create(ctx context.Context, params *dto.CreateSKURequest) (*models.SKU, error) {
	if params.Unit == "" {
		params.Unit = "cái"
	}

	result, err := uc.skuRepo.Create(params)
	if err != nil {
		uc.logger.Errorf("Create SKU err: %v", err)
		return nil, err
	}
	return result, nil
}
