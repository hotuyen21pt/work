package auth

import (
	"github.com/gin-gonic/gin"

	"lot-control/internal/config"
	"lot-control/internal/middleware"
	httphandler "lot-control/internal/modules/auth/delivery/http-handler"
	"lot-control/internal/modules/auth/usecase"
)

func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config, uc usecase.IUseCase) {
	h := httphandler.InitHandler(cfg, uc)
	grp := rg.Group("/auth")
	{
		grp.POST("/login", h.Login)
		grp.GET("/me", middleware.AuthMiddleware(cfg, ""), h.Me)
	}
}
