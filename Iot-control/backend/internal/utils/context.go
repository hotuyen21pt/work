// Package utils chứa các helper dùng chung, đặc biệt là đọc thông tin
// người dùng đã đăng nhập từ gin.Context (do AuthMiddleware nạp vào).
package utils

import (
	"github.com/gin-gonic/gin"

	"lot-control/internal/models"
)

// ContextClaimsKey là khóa lưu *models.Claims trong gin.Context.
const ContextClaimsKey = "claims"

// GetClaimsFromContext trả về claims của người dùng đã đăng nhập, hoặc nil.
func GetClaimsFromContext(ctx *gin.Context) *models.Claims {
	v, ok := ctx.Get(ContextClaimsKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*models.Claims)
	return claims
}

// GetLoggedUserIDFromContext trả về ID người dùng đã đăng nhập, 0 nếu không có.
func GetLoggedUserIDFromContext(ctx *gin.Context) int64 {
	if claims := GetClaimsFromContext(ctx); claims != nil {
		return claims.UserID
	}
	return 0
}

// GetLoggedUserBranchFromContext trả về chi nhánh của người dùng đã đăng nhập.
func GetLoggedUserBranchFromContext(ctx *gin.Context) string {
	if claims := GetClaimsFromContext(ctx); claims != nil {
		return claims.Branch
	}
	return ""
}

// GetLoggedUserRoleFromContext trả về vai trò của người dùng đã đăng nhập.
func GetLoggedUserRoleFromContext(ctx *gin.Context) string {
	if claims := GetClaimsFromContext(ctx); claims != nil {
		return claims.Role
	}
	return ""
}
