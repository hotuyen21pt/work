package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"lot-control/internal/config"
	"lot-control/internal/models"
	"lot-control/internal/utils"
	httperrors "lot-control/pkg/httperrors"
)

// AuthMiddleware xác thực JWT từ header Authorization và nạp claims vào context.
// requiredRole != "" sẽ bắt buộc người dùng phải có đúng vai trò đó (vd: "admin").
func AuthMiddleware(cfg *config.Config, requiredRole string) gin.HandlerFunc {
	secret := cfg.Server.JWTSecret
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httperrors.ResponseError{Message: "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httperrors.ResponseError{Message: "invalid authorization header"})
			return
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httperrors.ResponseError{Message: "invalid or expired token"})
			return
		}

		mc, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, httperrors.ResponseError{Message: "invalid claims"})
			return
		}

		claims := &models.Claims{
			UserID:   int64(mc["user_id"].(float64)),
			Username: mc["username"].(string),
			Branch:   mc["branch"].(string),
			Role:     mc["role"].(string),
		}

		if requiredRole != "" && claims.Role != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, httperrors.ResponseError{Message: "không đủ quyền truy cập"})
			return
		}

		c.Set(utils.ContextClaimsKey, claims)
		c.Next()
	}
}
