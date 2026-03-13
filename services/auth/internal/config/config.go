package config

import (
	"log"
	"os"
	"time"

	sharedconfig "github.com/animalpride/animalpride-core/services/shared/config"
	"gopkg.in/yaml.v3"
)

type Server struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}
type Database struct {
	Host     string       `yaml:"host"`
	Port     int          `yaml:"port"`
	User     string       `yaml:"user"`
	Password string       `yaml:"password"`
	DBName   string       `yaml:"authdbname"`
	Pool     DatabasePool `yaml:"pool"`
}

// DatabasePool is the shared connection-pool configuration for all services.
type DatabasePool = sharedconfig.DatabasePool

type Email struct {
	SMTPHost      string `yaml:"smtp_host"`
	SMTPPort      int    `yaml:"smtp_port"`
	SMTPTLS       bool   `yaml:"smtp_tls"`
	SMTPAuth      bool   `yaml:"smtp_auth"`
	SMTPUser      string `yaml:"smtp_user"`
	SMTPPassword  string `yaml:"smtp_password"`
	FromEmail     string `yaml:"from_email"`
	FromName      string `yaml:"from_name"`
	InviteBaseURL string `yaml:"invite_base_url"`
	ResetBaseURL  string `yaml:"reset_base_url"`
}

type AuthSession struct {
	AccessTokenTTL          time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL         time.Duration `yaml:"refresh_token_ttl"`
	RefreshRotationInterval time.Duration `yaml:"refresh_rotation_interval"`
}

type Config struct {
	Server      Server      `yaml:"server"`
	Database    Database    `yaml:"database"`
	Email       Email       `yaml:"email"`
	JWTSecret   string      `yaml:"jwt_secret"`
	AuthSession AuthSession `yaml:"auth_session"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("LoadConfig: open failed: %v", err)
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		log.Printf("LoadConfig: decode failed: %v", err)
		return nil, err
	}

	// Pull secrets from Docker/Kubernetes secret mounts/env, fall back to config file.
	cfg.Database.Password = sharedconfig.ResolveSecret("PARTNERS_AUTH_DB_PASSWORD", cfg.Database.Password)
	cfg.JWTSecret = sharedconfig.ResolveSecret("PARTNERS_AUTH_JWT_SECRET", cfg.JWTSecret)
	cfg.Email.SMTPPassword = sharedconfig.ResolveSecret("PARTNERS_AUTH_SMTP_PASSWORD", cfg.Email.SMTPPassword)
	if cfg.AuthSession.AccessTokenTTL == 0 {
		cfg.AuthSession.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.AuthSession.RefreshTokenTTL == 0 {
		cfg.AuthSession.RefreshTokenTTL = 30 * 24 * time.Hour
	}
	if cfg.AuthSession.RefreshRotationInterval == 0 {
		cfg.AuthSession.RefreshRotationInterval = 24 * time.Hour
	}
	return &cfg, nil
}
