package httphandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lot-control/internal/modules/user/dto"
	httperrors "lot-control/pkg/httperrors"
)

func (handler *userHandler) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "id không hợp lệ"})
		return
	}

	data := &dto.UpdateUserRequest{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: err.Error()})
		return
	}

	if err := handler.uc.Update(ctx, id, data); err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "đã cập nhật"})
}
