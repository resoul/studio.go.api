package usecase

type RegisterDTO struct {
	FullName string
	Email    string
	Password string
}

type LoginDTO struct {
	Email    string
	Password string
}

type VerifyEmailDTO struct {
	Email string
	Code  string
}

type ResetPasswordRequestDTO struct {
	Email string
}

type ResetPasswordDTO struct {
	Email       string
	Code        string
	NewPassword string
}

type UserDTO struct {
	ID              uint
	UUID            string
	FullName        string
	Email           string
	EmailVerified   bool
	EmailVerifiedAt *int64
	CreatedAt       int64
	UpdatedAt       int64
}
