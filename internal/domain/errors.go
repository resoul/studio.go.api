package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrInvalidCode        = errors.New("invalid code")
	ErrCodeExpired        = errors.New("code expired")
)
