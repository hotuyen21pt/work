package user

import (
	"github.com/gin-gonic/gin"

	"lot-control/internal/config"
	"lot-control/internal/middleware"
	httphandler "lot-control/internal/modules/user/delivery/http-handler"
	"lot-control/internal/modules/user/usecase"
)

func RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config, uc usecase.IUseCase) {
	h := httphandler.InitHandler(cfg, uc)
	// Toàn bộ route quản lý tài khoản yêu cầu vai trò admin.
	grp := rg.Group("/users", middleware.AuthMiddleware(cfg, "admin"))
	{
		grp.GET("", h.List)
		grp.POST("", h.Create)
		grp.PUT("/:id", h.Update)
		grp.DELETE("/:id", h.Delete)
	}
}
