package usecase

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/auth/repository"
	"lot-control/pkg/logger"
)

type authUseCase struct {
	cfg      *config.Config
	logger   logger.ILogger
	authRepo repository.IAuthRepository
}

func InitUseCase(cfg *config.Config, log logger.ILogger, authRepo repository.IAuthRepository) IUseCase {
	return &authUseCase{cfg: cfg, logger: log, authRepo: authRepo}
}
