package httphandler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	httperrors "lot-control/pkg/httperrors"
)

func (handler *lotHandler) List(ctx *gin.Context) {
	skuIDStr := ctx.Query("sku_id")
	if skuIDStr == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "sku_id là bắt buộc"})
		return
	}
	skuID, err := strconv.ParseInt(skuIDStr, 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "sku_id không hợp lệ"})
		return
	}

	res, err := handler.uc.ListBySKU(ctx, skuID)
	if err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}
