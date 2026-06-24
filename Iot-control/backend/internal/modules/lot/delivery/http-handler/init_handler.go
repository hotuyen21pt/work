package httphandler

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/lot/usecase"
)

type lotHandler struct {
	cfg *config.Config
	uc  usecase.IUseCase
}

func InitHandler(cfg *config.Config, uc usecase.IUseCase) *lotHandler {
	return &lotHandler{cfg: cfg, uc: uc}
}
