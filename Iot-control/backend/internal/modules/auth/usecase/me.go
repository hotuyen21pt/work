package usecase

import (
	"context"

	"lot-control/internal/models"
)

func (uc *authUseCase) Me(ctx context.Context, userID int64) (*models.User, error) {
	user, err := uc.authRepo.GetUserByID(userID)
	if err != nil {
		uc.logger.Errorf("Me GetUserByID err: %v", err)
		return nil, err
	}
	return user, nil
}
