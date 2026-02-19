package data

import "time"

type UserModel struct {
	ID                     uint       `gorm:"primaryKey;comment:ID"`
	Username               string     `gorm:"type:varchar(100);uniqueIndex;not null;comment:Username"`
	FullName               string     `gorm:"type:varchar(255);not null;default:'';comment:Full Name"`
	Email                  string     `gorm:"type:varchar(255);uniqueIndex;not null;comment:Email"`
	PasswordHash           string     `gorm:"type:varchar(255);not null;comment:Password Hash"`
	EmailVerifiedAt        *time.Time `gorm:"comment:Email Verified At"`
	VerificationCode       string     `gorm:"type:varchar(20);comment:Verification Code"`
	VerificationExpiresAt  *time.Time `gorm:"comment:Verification Code Expires At"`
	ResetPasswordCode      string     `gorm:"type:varchar(20);comment:Reset Password Code"`
	ResetPasswordExpiresAt *time.Time `gorm:"comment:Reset Password Code Expires At"`
	CreatedAt              time.Time  `gorm:"autoCreateTime;comment:Created At"`
	UpdatedAt              time.Time  `gorm:"autoUpdateTime;comment:Updated At"`
}

func (UserModel) TableName() string {
	return "users"
}
