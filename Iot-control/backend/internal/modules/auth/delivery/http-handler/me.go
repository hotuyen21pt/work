package httphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lot-control/internal/utils"
	httperrors "lot-control/pkg/httperrors"
)

func (handler *authHandler) Me(ctx *gin.Context) {
	userID := utils.GetLoggedUserIDFromContext(ctx)

	res, err := handler.uc.Me(ctx, userID)
	if err != nil {
		ctx.AbortWithStatusJSON(httperrors.GetStatusCode(err), httperrors.ResponseError{Message: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}
