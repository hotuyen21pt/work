package usecase

import (
	"context"

	"lot-control/internal/models"
)

func (uc *skuUseCase) List(ctx context.Context, q string) ([]models.SKU, error) {
	result, err := uc.skuRepo.List(q)
	if err != nil {
		uc.logger.Errorf("List SKU err: %v", err)
		return nil, err
	}
	return result, nil
}
