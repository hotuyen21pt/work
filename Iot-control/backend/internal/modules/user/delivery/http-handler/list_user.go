package httphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httperrors "lot-control/pkg/httperrors"
)

func (handler *userHandler) List(ctx *gin.Context) {
	res, err := handler.uc.List(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}
