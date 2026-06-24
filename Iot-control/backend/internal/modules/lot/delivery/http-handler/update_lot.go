package httphandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lot-control/internal/modules/lot/dto"
	"lot-control/internal/utils"
	httperrors "lot-control/pkg/httperrors"
)

func (handler *lotHandler) Update(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "id không hợp lệ"})
		return
	}

	data := &dto.UpdateLotRequest{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: err.Error()})
		return
	}

	userID := utils.GetLoggedUserIDFromContext(ctx)
	data.CountedByID = &userID

	res, err := handler.uc.Update(ctx, id, data)
	if err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}
