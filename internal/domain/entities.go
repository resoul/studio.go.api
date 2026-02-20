package domain

import "time"

type User struct {
	ID                     uint
	UUID                   string
	FullName               string
	Email                  string
	PasswordHash           string
	RegistrationIP         string
	RegistrationUserAgent  string
	LoginCount             uint
	EmailVerifiedAt        *time.Time
	VerificationCode       string
	VerificationExpiresAt  *time.Time
	ResetPasswordCode      string
	ResetPasswordExpiresAt *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
