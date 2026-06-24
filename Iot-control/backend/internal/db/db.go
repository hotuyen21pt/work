package db

import (
	stdsql "database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"lot-control/internal/config"
	"lot-control/internal/models"
)

// Init đảm bảo database tồn tại, kết nối GORM, chạy migrate và seed.
func Init(cfg *config.Config) (*gorm.DB, error) {
	d := cfg.Database

	// 1. Tạo database nếu chưa có (kết nối không kèm tên DB).
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=true",
		d.User, d.Password, d.Host, d.Port)
	rootDB, err := stdsql.Open("mysql", rootDSN)
	if err != nil {
		return nil, fmt.Errorf("kết nối MySQL: %w", err)
	}
	if err := rootDB.Ping(); err != nil {
		rootDB.Close()
		return nil, fmt.Errorf("không ping được MySQL (kiểm tra server/thông tin trong config.yaml): %w", err)
	}
	_, err = rootDB.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", d.Name))
	rootDB.Close()
	if err != nil {
		return nil, fmt.Errorf("tạo database %q: %w", d.Name, err)
	}

	// 2. Kết nối GORM tới database.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("GORM mở kết nối: %w", err)
	}

	// 3. Tạo/cập nhật bảng.
	if err := gdb.AutoMigrate(&models.User{}, &models.SKU{}, &models.Lot{}); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}

	// 4. Tạo tài khoản admin mặc định nếu chưa có.
	if err := seed(gdb); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	return gdb, nil
}

func seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Username:     "admin",
		PasswordHash: string(hash),
		FullName:     "Administrator",
		Branch:       "HQ",
		Role:         "admin",
	}
	return db.Create(&admin).Error
}
