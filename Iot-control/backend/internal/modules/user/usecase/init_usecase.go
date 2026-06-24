package usecase

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/user/repository"
	"lot-control/pkg/logger"
)

type userUseCase struct {
	cfg      *config.Config
	logger   logger.ILogger
	userRepo repository.IUserRepository
}

func InitUseCase(cfg *config.Config, log logger.ILogger, userRepo repository.IUserRepository) IUseCase {
	return &userUseCase{cfg: cfg, logger: log, userRepo: userRepo}
}
