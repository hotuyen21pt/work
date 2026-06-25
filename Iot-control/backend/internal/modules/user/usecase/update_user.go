package usecase

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"lot-control/internal/modules/user/dto"
	httperrors "lot-control/pkg/httperrors"
)

func (uc *userUseCase) Update(ctx context.Context, id int64, params *dto.UpdateUserRequest) error {
	// Kiểm tra tồn tại trước (trả 404 nếu không có) — username không cho đổi.
	if _, err := uc.userRepo.GetByID(id); err != nil {
		uc.logger.Errorf("Update user GetByID err: %v", err)
		return err
	}

	if params.Role == "" {
		params.Role = "staff"
	}

	fields := map[string]interface{}{
		"full_name": params.FullName,
		"branch":    params.Branch,
		"role":      params.Role,
	}

	// Chỉ đổi mật khẩu khi người dùng nhập giá trị mới.
	if params.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
		if err != nil {
			uc.logger.Errorf("Update user hash password err: %v", err)
			return httperrors.NewInternal("không thể hash mật khẩu")
		}
		fields["password_hash"] = string(hash)
	}

	if err := uc.userRepo.Update(id, fields); err != nil {
		uc.logger.Errorf("Update user err: %v", err)
		return err
	}
	return nil
}
