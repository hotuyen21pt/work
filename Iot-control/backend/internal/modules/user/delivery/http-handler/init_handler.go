package httphandler

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/user/usecase"
)

type userHandler struct {
	cfg *config.Config
	uc  usecase.IUseCase
}

func InitHandler(cfg *config.Config, uc usecase.IUseCase) *userHandler {
	return &userHandler{cfg: cfg, uc: uc}
}
