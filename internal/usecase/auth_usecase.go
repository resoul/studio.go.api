package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/football.manager.api/internal/domain"
	platformauth "github.com/football.manager.api/internal/platform/auth"
	"github.com/football.manager.api/internal/platform/mailer"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(ctx context.Context, dto RegisterDTO, ip, userAgent, locale string) (*UserDTO, error)
	VerifyEmail(ctx context.Context, dto VerifyEmailDTO, locale string) (*UserDTO, string, bool, error)
	ResendVerificationCode(ctx context.Context, dto ResendVerificationDTO, locale string) error
	Login(ctx context.Context, dto LoginDTO, ip, userAgent string) (*UserDTO, string, bool, error)
	AdminLogin(ctx context.Context, dto LoginDTO, ip, userAgent string) (*UserDTO, string, bool, error)
	RequestPasswordReset(ctx context.Context, dto ResetPasswordRequestDTO, locale string) error
	ResetPassword(ctx context.Context, dto ResetPasswordDTO, locale string) error
}

type authUseCase struct {
	userRepo          domain.UserRepository
	managerRepo       domain.ManagerRepository
	tokenMngr         *platformauth.UserTokenManager
	emailer           mailer.EmailSender
	notifyAdminEmails []string
}

func NewAuthUseCase(
	userRepo domain.UserRepository,
	managerRepo domain.ManagerRepository,
	tokenMngr *platformauth.UserTokenManager,
	emailer mailer.EmailSender,
	notifyAdminEmails []string,
) AuthUseCase {
	return &authUseCase{
		userRepo:          userRepo,
		managerRepo:       managerRepo,
		tokenMngr:         tokenMngr,
		emailer:           emailer,
		notifyAdminEmails: notifyAdminEmails,
	}
}

func (uc *authUseCase) Register(ctx context.Context, dto RegisterDTO, ip, userAgent, locale string) (*UserDTO, error) {
	email := strings.TrimSpace(strings.ToLower(dto.Email))
	fullName := strings.TrimSpace(dto.FullName)
	if email == "" || dto.Password == "" || fullName == "" {
		return nil, fmt.Errorf("full name, email and password are required")
	}

	existing, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
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
		UUID:                  uuid.New().String(),
		FullName:              fullName,
		Email:                 email,
		PasswordHash:          string(passwordHash),
		RegistrationIP:        sanitizeIP(ip),
		RegistrationUserAgent: sanitizeUserAgent(userAgent),
		VerificationCode:      code,
		VerificationExpiresAt: &expiresAt,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	verifyEmail, err := registrationCodeEmail(locale, user.FullName, code, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to render registration email: %w", err)
	}
	if err := uc.emailer.Send(ctx, user.Email, verifyEmail.Subject, verifyEmail.TextBody, verifyEmail.HTMLBody); err != nil {
		return nil, fmt.Errorf("failed to send verification code: %w", err)
	}
	if err := uc.notifyAdminsAboutRegistration(ctx, user, ip, userAgent, locale); err != nil {
		return nil, err
	}

	return mapUserToDTO(user), nil
}

func (uc *authUseCase) VerifyEmail(ctx context.Context, dto VerifyEmailDTO, locale string) (*UserDTO, string, bool, error) {
	user, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(dto.Email)))
	if err != nil {
		return nil, "", false, err
	}

	if user.VerificationCode != dto.Code {
		return nil, "", false, domain.ErrInvalidCode
	}

	if user.VerificationExpiresAt == nil || time.Now().UTC().After(*user.VerificationExpiresAt) {
		return nil, "", false, domain.ErrCodeExpired
	}

	now := time.Now().UTC()
	user.EmailVerifiedAt = &now
	user.VerificationCode = ""
	user.VerificationExpiresAt = nil
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, "", false, err
	}

	successEmail, err := emailVerifiedSuccessEmail(locale, user.FullName)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to render verified email: %w", err)
	}
	if err := uc.emailer.Send(ctx, user.Email, successEmail.Subject, successEmail.TextBody, successEmail.HTMLBody); err != nil {
		return nil, "", false, err
	}

	tokenRole := platformauth.RoleUser
	if user.Role == platformauth.RoleAdmin {
		tokenRole = platformauth.RoleAdmin
	}

	token, err := uc.tokenMngr.Generate(user.ID, tokenRole)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to generate token: %w", err)
	}

	managerExists, err := uc.managerRepo.ExistsByUserID(ctx, user.ID)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to check onboarding status: %w", err)
	}

	return mapUserToDTO(user), token, !managerExists, nil
}

func (uc *authUseCase) ResendVerificationCode(ctx context.Context, dto ResendVerificationDTO, locale string) error {
	user, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(dto.Email)))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil
		}
		return err
	}

	if user.EmailVerifiedAt != nil {
		return nil
	}

	code, err := generateCode(6)
	if err != nil {
		return err
	}
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	user.VerificationCode = code
	user.VerificationExpiresAt = &expiresAt

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	verifyEmail, err := registrationCodeEmail(locale, user.FullName, code, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to render registration email: %w", err)
	}

	return uc.emailer.Send(ctx, user.Email, verifyEmail.Subject, verifyEmail.TextBody, verifyEmail.HTMLBody)
}

func (uc *authUseCase) Login(ctx context.Context, dto LoginDTO, ip, userAgent string) (*UserDTO, string, bool, error) {
	return uc.login(ctx, dto, ip, userAgent, false)
}

func (uc *authUseCase) AdminLogin(ctx context.Context, dto LoginDTO, ip, userAgent string) (*UserDTO, string, bool, error) {
	return uc.login(ctx, dto, ip, userAgent, true)
}

func (uc *authUseCase) login(ctx context.Context, dto LoginDTO, ip, userAgent string, requireAdmin bool) (*UserDTO, string, bool, error) {
	user, err := uc.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(dto.Email)))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, "", false, domain.ErrInvalidCredentials
		}
		return nil, "", false, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		return nil, "", false, domain.ErrInvalidCredentials
	}

	if user.EmailVerifiedAt == nil {
		return nil, "", false, domain.ErrEmailNotVerified
	}

	tokenRole := platformauth.RoleUser
	if user.Role == platformauth.RoleAdmin {
		tokenRole = platformauth.RoleAdmin
	}

	if requireAdmin && tokenRole != platformauth.RoleAdmin {
		return nil, "", false, domain.ErrAdminAccessDenied
	}

	token, err := uc.tokenMngr.Generate(user.ID, tokenRole)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to generate token: %w", err)
	}

	user.LoginCount++
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, "", false, fmt.Errorf("failed to update login metadata: %w", err)
	}
	if err := uc.userRepo.UpdateLastLogin(ctx, user.ID, sanitizeIP(ip), sanitizeUserAgent(userAgent)); err != nil {
		return nil, "", false, fmt.Errorf("failed to save last login data: %w", err)
	}

	managerExists, err := uc.managerRepo.ExistsByUserID(ctx, user.ID)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to check onboarding status: %w", err)
	}

	return mapUserToDTO(user), token, !managerExists, nil
}

func (uc *authUseCase) RequestPasswordReset(ctx context.Context, dto ResetPasswordRequestDTO, locale string) error {
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

	resetEmail, err := resetPasswordCodeEmail(locale, user.FullName, code, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to render reset email: %w", err)
	}
	return uc.emailer.Send(ctx, user.Email, resetEmail.Subject, resetEmail.TextBody, resetEmail.HTMLBody)
}

func (uc *authUseCase) ResetPassword(ctx context.Context, dto ResetPasswordDTO, locale string) error {
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

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	successEmail, err := passwordChangedSuccessEmail(locale, user.FullName)
	if err != nil {
		return fmt.Errorf("failed to render password changed email: %w", err)
	}
	return uc.emailer.Send(ctx, user.Email, successEmail.Subject, successEmail.TextBody, successEmail.HTMLBody)
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
		UUID:            user.UUID,
		FullName:        user.FullName,
		Email:           user.Email,
		EmailVerified:   user.EmailVerifiedAt != nil,
		EmailVerifiedAt: verifiedAt,
		CreatedAt:       user.CreatedAt.Unix(),
		UpdatedAt:       user.UpdatedAt.Unix(),
	}
}

func sanitizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if len(ip) > 45 {
		return ip[:45]
	}
	return ip
}

func sanitizeUserAgent(userAgent string) string {
	userAgent = strings.TrimSpace(userAgent)
	if len(userAgent) > 512 {
		return userAgent[:512]
	}
	return userAgent
}

func (uc *authUseCase) notifyAdminsAboutRegistration(ctx context.Context, user *domain.User, ip, userAgent, locale string) error {
	if len(uc.notifyAdminEmails) == 0 {
		return nil
	}

	adminEmail, err := adminNewRegistrationEmail(locale, user.FullName, user.Email, sanitizeIP(ip), sanitizeUserAgent(userAgent), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to render admin email: %w", err)
	}
	for _, to := range uc.notifyAdminEmails {
		if err := uc.emailer.Send(ctx, to, adminEmail.Subject, adminEmail.TextBody, adminEmail.HTMLBody); err != nil {
			return fmt.Errorf("failed to notify admin %s: %w", to, err)
		}
	}

	return nil
}
