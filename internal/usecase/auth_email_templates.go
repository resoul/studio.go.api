package usecase

import (
	"fmt"
	"strings"
	"time"
)

type renderedEmail struct {
	Subject  string
	TextBody string
	HTMLBody string
}

type emailTemplateData struct {
	Subject      string
	DisplayName  string
	FullName     string
	Email        string
	Code         string
	ExpiresAtUTC string
	IP           string
	UserAgent    string
	CreatedAtUTC string
}

func registrationCodeEmail(locale, fullName, code string, expiresAt time.Time) (renderedEmail, error) {
	data := emailTemplateData{
		DisplayName:  displayName(fullName),
		FullName:     strings.TrimSpace(fullName),
		Code:         code,
		ExpiresAtUTC: expiresAt.UTC().Format(time.RFC3339),
	}
	return renderLocalizedEmail(locale, "registration_code", data)
}

func emailVerifiedSuccessEmail(locale, fullName string) (renderedEmail, error) {
	data := emailTemplateData{
		DisplayName: displayName(fullName),
		FullName:    strings.TrimSpace(fullName),
	}
	return renderLocalizedEmail(locale, "email_verified", data)
}

func resetPasswordCodeEmail(locale, fullName, code string, expiresAt time.Time) (renderedEmail, error) {
	data := emailTemplateData{
		DisplayName:  displayName(fullName),
		FullName:     strings.TrimSpace(fullName),
		Code:         code,
		ExpiresAtUTC: expiresAt.UTC().Format(time.RFC3339),
	}
	return renderLocalizedEmail(locale, "reset_password_code", data)
}

func passwordChangedSuccessEmail(locale, fullName string) (renderedEmail, error) {
	data := emailTemplateData{
		DisplayName: displayName(fullName),
		FullName:    strings.TrimSpace(fullName),
	}
	return renderLocalizedEmail(locale, "password_changed", data)
}

func adminNewRegistrationEmail(locale, fullName, email, ip, userAgent string, createdAt time.Time) (renderedEmail, error) {
	data := emailTemplateData{
		DisplayName:  displayName(fullName),
		FullName:     strings.TrimSpace(fullName),
		Email:        strings.TrimSpace(email),
		IP:           defaultValue(ip, "-"),
		UserAgent:    defaultValue(userAgent, "-"),
		CreatedAtUTC: createdAt.UTC().Format(time.RFC3339),
	}
	return renderLocalizedEmail(locale, "admin_new_registration", data)
}

func displayName(fullName string) string {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return "there"
	}
	return fullName
}

func defaultValue(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if strings.HasPrefix(locale, "ru") {
		return "ru"
	}
	return "en"
}

func templatePath(locale, name, part string) string {
	return fmt.Sprintf("templates/email/%s/%s.%s.tmpl", locale, name, part)
}
