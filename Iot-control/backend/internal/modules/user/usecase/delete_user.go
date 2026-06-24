package usecase

import (
	"context"

	httperrors "lot-control/pkg/httperrors"
)

func (uc *userUseCase) Delete(ctx context.Context, id int64) error {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		uc.logger.Errorf("Delete user GetByID err: %v", err)
		return err
	}
	if user.Role == "admin" {
		return httperrors.NewBadRequest("không thể xóa tài khoản admin")
	}

	if err := uc.userRepo.Delete(id); err != nil {
		uc.logger.Errorf("Delete user err: %v", err)
		return err
	}
	return nil
}
