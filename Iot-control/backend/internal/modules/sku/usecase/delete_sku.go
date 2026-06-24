package usecase

import (
	"context"
)

func (uc *skuUseCase) Delete(ctx context.Context, id int64) error {
	if err := uc.skuRepo.Delete(id); err != nil {
		uc.logger.Errorf("Delete SKU err: %v", err)
		return err
	}
	return nil
}
