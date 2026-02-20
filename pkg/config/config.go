package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	DB      DatabaseConfig `envconfig:"DB"`
	Logging LoggingConfig  `envconfig:"LOG"`
	Auth    AuthConfig     `envconfig:"AUTH"`
	Server  ServerConfig   `envconfig:"SERVER"`
	Mailer  MailerConfig   `envconfig:"MAILER"`
}

type DatabaseConfig struct {
	DSN string `envconfig:"DSN" required:"true"`
}

type LoggingConfig struct {
	Level string `envconfig:"LEVEL" default:"debug"`
}

type AuthConfig struct {
	APITokens     string `envconfig:"API_TOKENS" required:"false"`
	JWTSecret     string `envconfig:"JWT_SECRET" default:"dev-jwt-secret-change-me"`
	JWTTTLMinutes int    `envconfig:"JWT_TTL_MINUTES" default:"60"`
}

type ServerConfig struct {
	Port               string `envconfig:"PORT" default:"8080"`
	CORSAllowedOrigins string `envconfig:"CORS_ALLOWED_ORIGINS" default:"http://dashboard.manager.localhost,http://localhost:5173"`
}

type MailerConfig struct {
	Provider    string `envconfig:"PROVIDER" default:"log"`
	From        string `envconfig:"FROM" default:"no-reply@manager.localhost"`
	Host        string `envconfig:"HOST" default:"localhost"`
	Port        int    `envconfig:"PORT" default:"1025"`
	Username    string `envconfig:"USERNAME" required:"false"`
	Password    string `envconfig:"PASSWORD" required:"false"`
	LogoPath    string `envconfig:"LOGO_PATH" default:"logo.png"`
	AdminEmails string `envconfig:"ADMIN_EMAILS" required:"false"`
}

func (s *ServerConfig) GetCORSAllowedOrigins() []string {
	origins := strings.Split(s.CORSAllowedOrigins, ",")
	result := make([]string, 0, len(origins))

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			result = append(result, origin)
		}
	}

	return result
}

func (a *AuthConfig) GetTokens() []string {
	if a.APITokens == "" {
		return []string{}
	}

	tokens := strings.Split(a.APITokens, ",")
	result := make([]string, 0, len(tokens))

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token != "" {
			result = append(result, token)
		}
	}

	return result
}

func (m *MailerConfig) GetAdminEmails() []string {
	if m.AdminEmails == "" {
		return []string{}
	}

	emails := strings.Split(m.AdminEmails, ",")
	result := make([]string, 0, len(emails))
	for _, email := range emails {
		email = strings.TrimSpace(strings.ToLower(email))
		if email != "" {
			result = append(result, email)
		}
	}

	return result
}

func Init(ctx context.Context) *Config {
	cfg, err := loadConfig(ctx)
	if err != nil {
		panic(err)
	}

	return cfg
}

func loadConfig(ctx context.Context) (*Config, error) {
	if ctx == nil {
		panic("context must not be nil")
	}

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}
	applyLegacyEmailFallback(&cfg.Mailer)

	return &cfg, nil
}

func applyLegacyEmailFallback(mailer *MailerConfig) {
	if mailer == nil {
		return
	}

	if v := strings.TrimSpace(os.Getenv("EMAIL_PROVIDER")); v != "" && os.Getenv("MAILER_PROVIDER") == "" {
		mailer.Provider = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_FROM")); v != "" && os.Getenv("MAILER_FROM") == "" {
		mailer.From = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_HOST")); v != "" && os.Getenv("MAILER_HOST") == "" {
		mailer.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_PORT")); v != "" && os.Getenv("MAILER_PORT") == "" {
		var port int
		if _, err := fmt.Sscanf(v, "%d", &port); err == nil {
			mailer.Port = port
		}
	}
	if v := os.Getenv("EMAIL_USERNAME"); v != "" && os.Getenv("MAILER_USERNAME") == "" {
		mailer.Username = v
	}
	if v := os.Getenv("EMAIL_PASSWORD"); v != "" && os.Getenv("MAILER_PASSWORD") == "" {
		mailer.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("EMAIL_ADMIN_EMAILS")); v != "" && os.Getenv("MAILER_ADMIN_EMAILS") == "" {
		mailer.AdminEmails = v
	}
}
