package usecase

import (
	"context"

	"lot-control/internal/models"
	"lot-control/internal/modules/user/dto"
)

type IUseCase interface {
	List(ctx context.Context) ([]models.User, error)
	Create(ctx context.Context, params *dto.CreateUserRequest) (*models.User, error)
	Delete(ctx context.Context, id int64) error
}
