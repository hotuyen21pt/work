package httphandler

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/sku/usecase"
)

type skuHandler struct {
	cfg *config.Config
	uc  usecase.IUseCase
}

func InitHandler(cfg *config.Config, uc usecase.IUseCase) *skuHandler {
	return &skuHandler{cfg: cfg, uc: uc}
}
