package usecase

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"lot-control/internal/modules/auth/dto"
	httperrors "lot-control/pkg/httperrors"
)

func (uc *authUseCase) Login(ctx context.Context, params *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := uc.authRepo.GetUserByUsername(params.Username)
	if err != nil {
		uc.logger.Errorf("Login GetUserByUsername err: %v", err)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.Password)); err != nil {
		return nil, httperrors.NewUnauthorized("Tên đăng nhập hoặc mật khẩu không đúng")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"branch":   user.Branch,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(uc.cfg.Server.JWTSecret))
	if err != nil {
		uc.logger.Errorf("Login sign token err: %v", err)
		return nil, httperrors.NewInternal("không thể tạo token")
	}

	return &dto.LoginResponse{Token: tokenStr, User: *user}, nil
}
