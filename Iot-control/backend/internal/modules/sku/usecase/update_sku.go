package usecase

import (
	"context"

	"lot-control/internal/modules/sku/dto"
)

func (uc *skuUseCase) Update(ctx context.Context, id int64, params *dto.UpdateSKURequest) error {
	if params.Unit == "" {
		params.Unit = "cái"
	}

	if err := uc.skuRepo.Update(id, params); err != nil {
		uc.logger.Errorf("Update SKU err: %v", err)
		return err
	}
	return nil
}
