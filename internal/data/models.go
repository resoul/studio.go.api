package data

import "time"

type UserModel struct {
	ID                     uint       `gorm:"primaryKey;comment:ID"`
	UUID                   string     `gorm:"type:char(36);uniqueIndex;not null;column:uuid;comment:UUID"`
	FullName               string     `gorm:"type:varchar(255);not null;default:'';comment:Full Name"`
	Email                  string     `gorm:"type:varchar(255);uniqueIndex;not null;comment:Email"`
	PasswordHash           string     `gorm:"type:varchar(255);not null;comment:Password Hash"`
	Role                   string     `gorm:"type:varchar(20);not null;default:'user';comment:Role"`
	RegistrationIP         string     `gorm:"type:varchar(45);not null;default:'';comment:Registration IP"`
	RegistrationUserAgent  string     `gorm:"type:varchar(512);not null;default:'';comment:Registration User Agent"`
	LoginCount             uint       `gorm:"not null;default:0;comment:Login Count"`
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

type UserLastLoginModel struct {
	UserID             uint       `gorm:"primaryKey;comment:User ID"`
	LastLoginAt        *time.Time `gorm:"comment:Last Login At"`
	LastLoginIP        string     `gorm:"type:varchar(45);not null;default:'';comment:Last Login IP"`
	LastLoginUserAgent string     `gorm:"type:varchar(512);not null;default:'';comment:Last Login User Agent"`
	CreatedAt          time.Time  `gorm:"autoCreateTime;comment:Created At"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime;comment:Updated At"`
}

func (UserLastLoginModel) TableName() string {
	return "user_last_logins"
}

type ManagerModel struct {
	ID        uint      `gorm:"primaryKey;comment:ID"`
	UserID    uint      `gorm:"uniqueIndex;not null;comment:User ID"`
	FirstName string    `gorm:"type:varchar(100);not null;comment:First Name"`
	LastName  string    `gorm:"type:varchar(100);not null;comment:Last Name"`
	Birthday  time.Time `gorm:"type:date;not null;comment:Birthday"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:Created At"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:Updated At"`
}

func (ManagerModel) TableName() string {
	return "managers"
}

type CareerModel struct {
	ID        uint      `gorm:"primaryKey;comment:ID"`
	ManagerID uint      `gorm:"index;not null;comment:Manager ID"`
	Name      string    `gorm:"type:varchar(160);not null;comment:Career Name"`
	CreatedAt time.Time `gorm:"autoCreateTime;comment:Created At"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;comment:Updated At"`
}

func (CareerModel) TableName() string {
	return "careers"
}
