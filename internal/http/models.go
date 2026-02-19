package http

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ConfirmResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type AuthSuccessResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user,omitempty"`
	Token   string       `json:"token,omitempty"`
}

type UserResponse struct {
	ID              uint   `json:"id"`
	Username        string `json:"username"`
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	EmailVerifiedAt *int64 `json:"email_verified_at,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}
