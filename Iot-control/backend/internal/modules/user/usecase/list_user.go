package usecase

import (
	"context"

	"lot-control/internal/models"
)

func (uc *userUseCase) List(ctx context.Context) ([]models.User, error) {
	result, err := uc.userRepo.List()
	if err != nil {
		uc.logger.Errorf("List user err: %v", err)
		return nil, err
	}
	return result, nil
}
