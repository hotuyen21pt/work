package usecase

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/lot/repository"
	"lot-control/pkg/logger"
)

type lotUseCase struct {
	cfg     *config.Config
	logger  logger.ILogger
	lotRepo repository.ILotRepository
}

func InitUseCase(cfg *config.Config, log logger.ILogger, lotRepo repository.ILotRepository) IUseCase {
	return &lotUseCase{cfg: cfg, logger: log, lotRepo: lotRepo}
}
