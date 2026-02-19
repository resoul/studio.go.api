package config

import (
	"context"
	"fmt"
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
	Email   EmailConfig    `envconfig:"EMAIL"`
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

type EmailConfig struct {
	Provider string `envconfig:"PROVIDER" default:"log"`
	From     string `envconfig:"FROM" default:"no-reply@manager.localhost"`
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     int    `envconfig:"PORT" default:"1025"`
	Username string `envconfig:"USERNAME" required:"false"`
	Password string `envconfig:"PASSWORD" required:"false"`
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

	return &cfg, nil
}
