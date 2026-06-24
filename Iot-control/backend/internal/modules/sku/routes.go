package sku

import (
	"github.com/gin-gonic/gin"

	"lot-control/internal/config"
	"lot-control/internal/middleware"
	httphandler "lot-control/internal/modules/sku/delivery/http-handler"
	"lot-control/internal/modules/sku/usecase"
)

func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config, uc usecase.IUseCase) {
	h := httphandler.InitHandler(cfg, uc)
	grp := rg.Group("/skus")
	{
		grp.GET("", middleware.AuthMiddleware(cfg, ""), h.List)
		grp.POST("", middleware.AuthMiddleware(cfg, ""), h.Create)
		grp.GET("/:id", middleware.AuthMiddleware(cfg, ""), h.GetDetail)
		grp.PUT("/:id", middleware.AuthMiddleware(cfg, ""), h.Update)
		grp.DELETE("/:id", middleware.AuthMiddleware(cfg, ""), h.Delete)
	}
}
