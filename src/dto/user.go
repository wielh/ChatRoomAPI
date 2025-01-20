package dto

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=5,max=50"`
	Password string `json:"password" binding:"required,min=5,max=50"`
	Name     string `json:"name" binding:"required"`
	Birthday string `json:"birthday" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type UserRegisterResponse struct {
	ID uint64 `json:"id" binding:"required"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponse struct {
	ID       uint64 `json:"id" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type ResetPasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	NewPassword string `json:"newpassword" binding:"required"`
}

type GetUserInfoRequest struct {
	ID uint64
}

type GetUserInfoResponse struct {
	Id       uint64 `json:"id" binding:"required"`
	Username string `json:"username" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Birthday string `json:"birthday" binding:"required"`
	Email    string `json:"email" binding:"required"`
}
