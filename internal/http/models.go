package http

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
	Message            string       `json:"message"`
	User               UserResponse `json:"user,omitempty"`
	Token              string       `json:"token,omitempty"`
	Role               string       `json:"role,omitempty"`
	OnboardingRequired bool         `json:"onboarding_required,omitempty"`
}

type UserResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	FullName        string `json:"full_name"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	EmailVerifiedAt *int64 `json:"email_verified_at,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

type CreateManagerRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Birthday  string `json:"birthday" binding:"required"`
}

type ManagerResponse struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthday  string `json:"birthday"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
