package httphandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	httperrors "lot-control/pkg/httperrors"
)

// UpdateImageBoxes cập nhật box của một ảnh lô (mở lại chỉnh tay) và ghi/xoá
// nhãn dataset tương ứng.
func (handler *lotHandler) UpdateImageBoxes(ctx *gin.Context) {
	lotID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "id không hợp lệ"})
		return
	}
	imageID, err := strconv.ParseInt(ctx.Param("imageId"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "imageId không hợp lệ"})
		return
	}

	var body struct {
		Boxes  json.RawMessage `json:"boxes"`
		Edited bool            `json:"edited"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	if err := handler.uc.UpdateImageBoxes(ctx, lotID, imageID, body.Boxes, body.Edited); err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
