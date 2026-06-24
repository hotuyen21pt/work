package usecase

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"lot-control/internal/models"
	"lot-control/internal/modules/user/dto"
	httperrors "lot-control/pkg/httperrors"
)

func (uc *userUseCase) Create(ctx context.Context, params *dto.CreateUserRequest) (*models.User, error) {
	if params.Role == "" {
		params.Role = "staff"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Errorf("Create user hash password err: %v", err)
		return nil, httperrors.NewInternal("không thể hash mật khẩu")
	}

	user := &models.User{
		Username:     params.Username,
		PasswordHash: string(hash),
		FullName:     params.FullName,
		Branch:       params.Branch,
		Role:         params.Role,
	}
	if err := uc.userRepo.Create(user); err != nil {
		uc.logger.Errorf("Create user err: %v", err)
		return nil, err
	}
	return user, nil
}
