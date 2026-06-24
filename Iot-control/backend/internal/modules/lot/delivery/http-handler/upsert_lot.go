package httphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lot-control/internal/modules/lot/dto"
	"lot-control/internal/utils"
	httperrors "lot-control/pkg/httperrors"
)

func (handler *lotHandler) Upsert(ctx *gin.Context) {
	data := &dto.UpsertLotRequest{}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: err.Error()})
		return
	}

	// Người kiểm = người đang đăng nhập; chi nhánh mặc định theo chi nhánh của họ.
	userID := utils.GetLoggedUserIDFromContext(ctx)
	data.CountedByID = &userID
	if data.Branch == "" {
		data.Branch = utils.GetLoggedUserBranchFromContext(ctx)
	}

	res, err := handler.uc.Upsert(ctx, data)
	if err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}
