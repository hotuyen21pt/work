package usecase

import (
	"context"

	"lot-control/internal/models"
)

func (uc *lotUseCase) ListBySKU(ctx context.Context, skuID int64) ([]models.Lot, error) {
	result, err := uc.lotRepo.ListBySKU(skuID)
	if err != nil {
		uc.logger.Errorf("ListBySKU lot err: %v", err)
		return nil, err
	}
	return result, nil
}
