package usecase

import (
	"context"
)

func (uc *lotUseCase) Delete(ctx context.Context, id int64) error {
	if err := uc.lotRepo.Delete(id); err != nil {
		uc.logger.Errorf("Delete lot err: %v", err)
		return err
	}
	return nil
}
