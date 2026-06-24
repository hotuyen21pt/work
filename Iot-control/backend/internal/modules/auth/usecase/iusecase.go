package usecase

import (
	"context"

	"lot-control/internal/models"
	"lot-control/internal/modules/auth/dto"
)

type IUseCase interface {
	Login(ctx context.Context, params *dto.LoginRequest) (*dto.LoginResponse, error)
	Me(ctx context.Context, userID int64) (*models.User, error)
}
