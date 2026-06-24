package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"lot-control/internal/middleware"
	"lot-control/internal/models"
)

type AuthHandler struct {
	db     *sql.DB
	secret string
}

func NewAuthHandler(db *sql.DB, secret string) *AuthHandler {
	return &AuthHandler{db: db, secret: secret}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.QueryRow(
		`SELECT id, username, password_hash, full_name, branch, role, created_at FROM users WHERE username = ?`,
		req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.FullName, &user.Branch, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tên đăng nhập hoặc mật khẩu không đúng"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lỗi cơ sở dữ liệu"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tên đăng nhập hoặc mật khẩu không đúng"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"branch":   user.Branch,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(h.secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không thể tạo token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenStr, "user": user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	err := h.db.QueryRow(
		`SELECT id, username, full_name, branch, role, created_at FROM users WHERE id = ?`,
		claims.UserID,
	).Scan(&user.ID, &user.Username, &user.FullName, &user.Branch, &user.Role, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lỗi cơ sở dữ liệu"})
		return
	}

	c.JSON(http.StatusOK, user)
}
