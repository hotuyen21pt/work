package usecase

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/sku/repository"
	"lot-control/pkg/logger"
)

type skuUseCase struct {
	cfg     *config.Config
	logger  logger.ILogger
	skuRepo repository.ISKURepository
}

func InitUseCase(cfg *config.Config, log logger.ILogger, skuRepo repository.ISKURepository) IUseCase {
	return &skuUseCase{cfg: cfg, logger: log, skuRepo: skuRepo}
}
