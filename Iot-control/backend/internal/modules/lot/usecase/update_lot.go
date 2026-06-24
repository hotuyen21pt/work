package usecase

import (
	"context"
	"time"

	"lot-control/internal/models"
	"lot-control/internal/modules/lot/dto"
)

func (uc *lotUseCase) Update(ctx context.Context, id int64, params *dto.UpdateLotRequest) (*models.Lot, error) {
	fields := map[string]interface{}{
		"manufacture_date": params.ManufactureDate,
		"expiry_date":      params.ExpiryDate,
		"qty":              params.Qty,
		"notes":            params.Notes,
		"counted_by":       params.CountedByID,
		"counted_at":       time.Now(),
	}

	if err := uc.lotRepo.Update(id, fields); err != nil {
		uc.logger.Errorf("Update lot err: %v", err)
		return nil, err
	}

	result, err := uc.lotRepo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Update lot reload err: %v", err)
		return nil, err
	}
	return result, nil
}
