package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/football.manager.api/internal/domain"
	"github.com/football.manager.api/internal/infrastructure"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(ctx context.Context, dto RegisterDTO) (*UserDTO, error)
	VerifyEmail(ctx context.Context, dto VerifyEmailDTO) error
	Login(ctx context.Context, dto LoginDTO) (*UserDTO, error)
	RequestPasswordReset(ctx context.Context, dto ResetPasswordRequestDTO) error
	ResetPassword(ctx context.Context, dto ResetPasswordDTO) error
}

type authUseCase struct {
	userRepo domain.UserRepository
	emailer  infrastructure.EmailSender
}

var usernamePattern = regexp.MustCompile(`^[a-z0-9_.-]{3,32}$`)

func NewAuthUseCase(userRepo domain.UserRepository, emailer infrastructure.EmailSender) AuthUseCase {
	return &authUseCase{
		userRepo: userRepo,
		emailer:  emailer,
	}
}

func (uc *authUseCase) Register(ctx context.Context, dto RegisterDTO) (*UserDTO, error) {
	email := strings.TrimSpace(strings.ToLower(dto.Email))
	username := strings.TrimSpace(strings.ToLower(dto.Username))
	fullName := strings.TrimSpace(dto.FullName)
	if email == "" || dto.Password == "" || username == "" || fullName == "" {
		return nil, fmt.Errorf("username, full name, email and password are required")
	}
	if !usernamePattern.MatchString(username) {
		return nil, domain.ErrInvalidUsername
	}

	existing, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	existingByUsername, err := uc.userRepo.GetByUsername(ctx, username)
	if err == nil && existingByUsername != nil {
		return nil, domain.ErrUsernameTaken
	}
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check existing username: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	code, err := generateCode(6)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)

	user := &domain.User{
		Username:              username,
		FullName:              fullName,
		Email:                 email,
		PasswordHash:          string(passwordHash),
		VerificationCode:      code,
		VerificationExpiresAt: &expiresAt,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := uc.emailer.Send(ctx, user.Email, "Account verification code", fmt.Sprintf("Your verification code is: %s", code)); err != nil {
		return nil, fmt.Errorf("failed to send verification code: %w", err)
	}

	return mapUserToDTO(user), nil
}

func (uc *authUseCase) VerifyEmail(ctx context.Context, dto VerifyEmailDTO) error {
	user, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(dto.Email)))
	if err != nil {
		return err
	}

	if user.VerificationCode != dto.Code {
		return domain.ErrInvalidCode
	}

	if user.VerificationExpiresAt == nil || time.Now().UTC().After(*user.VerificationExpiresAt) {
		return domain.ErrCodeExpired
	}

	now := time.Now().UTC()
	user.EmailVerifiedAt = &now
	user.VerificationCode = ""
	user.VerificationExpiresAt = nil
	return uc.userRepo.Update(ctx, user)
}

func (uc *authUseCase) Login(ctx context.Context, dto LoginDTO) (*UserDTO, error) {
	user, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(dto.Email)))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.EmailVerifiedAt == nil {
		return nil, domain.ErrEmailNotVerified
	}

	return mapUserToDTO(user), nil
}

func (uc *authUseCase) RequestPasswordReset(ctx context.Context, dto ResetPasswordRequestDTO) error {
	user, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(dto.Email)))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil
		}
		return err
	}

	code, err := generateCode(6)
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	user.ResetPasswordCode = code
	user.ResetPasswordExpiresAt = &expiresAt

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return uc.emailer.Send(ctx, user.Email, "Reset password code", fmt.Sprintf("Your reset password code is: %s", code))
}

func (uc *authUseCase) ResetPassword(ctx context.Context, dto ResetPasswordDTO) error {
	user, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(dto.Email)))
	if err != nil {
		return err
	}

	if user.ResetPasswordCode != dto.Code {
		return domain.ErrInvalidCode
	}

	if user.ResetPasswordExpiresAt == nil || time.Now().UTC().After(*user.ResetPasswordExpiresAt) {
		return domain.ErrCodeExpired
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(passwordHash)
	user.ResetPasswordCode = ""
	user.ResetPasswordExpiresAt = nil

	return uc.userRepo.Update(ctx, user)
}

func generateCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid code length")
	}

	const digits = "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate code: %w", err)
		}
		code[i] = digits[n.Int64()]
	}

	return string(code), nil
}

func mapUserToDTO(user *domain.User) *UserDTO {
	var verifiedAt *int64
	if user.EmailVerifiedAt != nil {
		ts := user.EmailVerifiedAt.Unix()
		verifiedAt = &ts
	}

	return &UserDTO{
		ID:              user.ID,
		Username:        user.Username,
		FullName:        user.FullName,
		Email:           user.Email,
		EmailVerified:   user.EmailVerifiedAt != nil,
		EmailVerifiedAt: verifiedAt,
		CreatedAt:       user.CreatedAt.Unix(),
		UpdatedAt:       user.UpdatedAt.Unix(),
	}
}
