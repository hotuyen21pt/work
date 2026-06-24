package httphandler

import (
	"lot-control/internal/config"
	"lot-control/internal/modules/auth/usecase"
)

type authHandler struct {
	cfg *config.Config
	uc  usecase.IUseCase
}

func InitHandler(cfg *config.Config, uc usecase.IUseCase) *authHandler {
	return &authHandler{cfg: cfg, uc: uc}
}
