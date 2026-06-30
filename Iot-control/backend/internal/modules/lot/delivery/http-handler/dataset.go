package httphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httperrors "lot-control/pkg/httperrors"
)

// SaveDataset nhận ảnh gốc + nhãn (định dạng YOLO) và lưu vào kho dataset.
func (handler *lotHandler) SaveDataset(ctx *gin.Context) {
	image, err := ctx.FormFile("image")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, httperrors.ResponseError{Message: "thiếu ảnh: " + err.Error()})
		return
	}
	labels := ctx.PostForm("labels")

	if err := handler.uc.SaveDataset(ctx, image, labels); err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
