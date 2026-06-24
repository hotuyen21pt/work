package repository

import (
	"errors"

	"gorm.io/gorm"

	"lot-control/internal/models"
	httperrors "lot-control/pkg/httperrors"
)

type IAuthRepository interface {
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
}

type authRepository struct {
	db *gorm.DB
}

func InitAuthRepository(db *gorm.DB) IAuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, httperrors.NewUnauthorized("Tên đăng nhập hoặc mật khẩu không đúng")
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
