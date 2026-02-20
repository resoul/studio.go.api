package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrManagerNotFound    = errors.New("manager not found")
	ErrManagerExists      = errors.New("manager already exists")
	ErrCareerNotFound     = errors.New("career not found")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrAdminAccessDenied  = errors.New("admin access denied")
	ErrInvalidCode        = errors.New("invalid code")
	ErrCodeExpired        = errors.New("code expired")
)
