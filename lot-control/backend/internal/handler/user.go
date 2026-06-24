package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"lot-control/internal/middleware"
	"lot-control/internal/models"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) List(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "chỉ admin mới có quyền"})
		return
	}

	rows, err := h.db.Query(`SELECT id, username, full_name, branch, role, created_at FROM users ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Branch, &u.Role, &u.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, u)
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "chỉ admin mới có quyền"})
		return
	}

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Role == "" {
		req.Role = "staff"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không thể hash mật khẩu"})
		return
	}

	res, err := h.db.Exec(
		`INSERT INTO users (username, password_hash, full_name, branch, role) VALUES (?,?,?,?,?)`,
		req.Username, string(hash), req.FullName, req.Branch, req.Role,
	)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "tên đăng nhập đã tồn tại"})
		return
	}

	id, _ := res.LastInsertId()
	var u models.User
	h.db.QueryRow(
		`SELECT id, username, full_name, branch, role, created_at FROM users WHERE id=?`, id,
	).Scan(&u.ID, &u.Username, &u.FullName, &u.Branch, &u.Role, &u.CreatedAt)

	c.JSON(http.StatusCreated, u)
}

func (h *UserHandler) Delete(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "chỉ admin mới có quyền"})
		return
	}

	id := c.Param("id")
	var role string
	if err := h.db.QueryRow(`SELECT role FROM users WHERE id=?`, id).Scan(&role); err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy user"})
		return
	}
	if role == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "không thể xóa tài khoản admin"})
		return
	}

	h.db.Exec(`DELETE FROM users WHERE id=?`, id)
	c.JSON(http.StatusOK, gin.H{"message": "đã xóa"})
}
