package dto

type UpdateUserRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Branch   string `json:"branch" binding:"required"`
	Role     string `json:"role"`
	// Password để trống = giữ nguyên mật khẩu cũ; có giá trị = đặt lại mật khẩu.
	Password string `json:"password"`
}
