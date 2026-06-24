package httphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lot-control/internal/modules/user/dto"
	httperrors "lot-control/pkg/httperrors"
)

func (handler *userHandler) Create(ctx *gin.Context) {
	data := &dto.CreateUserRequest{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: err.Error()})
		return
	}

	res, err := handler.uc.Create(ctx, data)
	if err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, res)
}
