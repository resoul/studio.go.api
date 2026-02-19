package domain

import "time"

type User struct {
	ID                     uint
	Username               string
	FullName               string
	Email                  string
	PasswordHash           string
	EmailVerifiedAt        *time.Time
	VerificationCode       string
	VerificationExpiresAt  *time.Time
	ResetPasswordCode      string
	ResetPasswordExpiresAt *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
