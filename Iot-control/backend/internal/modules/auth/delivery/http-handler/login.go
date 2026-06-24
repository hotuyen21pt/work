package httphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lot-control/internal/modules/auth/dto"
	httperrors "lot-control/pkg/httperrors"
)

func (handler *authHandler) Login(ctx *gin.Context) {
	data := &dto.LoginRequest{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: err.Error()})
		return
	}

	res, err := handler.uc.Login(ctx, data)
	if err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}
