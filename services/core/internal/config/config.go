package config

import (
	"os"

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
	DBName   string       `yaml:"coredbname"`
	Pool     DatabasePool `yaml:"pool"`
}

type DatabasePool = sharedconfig.DatabasePool

type Auth struct {
	BaseURL string `yaml:"base_url"`
}

type Email struct {
	ResendAPIKey   string `yaml:"resend_api_key"`
	FromEmail      string `yaml:"from_email"`
	FromName       string `yaml:"from_name"`
	PartnerLeadsTo string `yaml:"partner_leads_to"`
}

type ComingSoon struct {
	Enabled              bool   `yaml:"enabled"`
	Message              string `yaml:"message"`
	PreviewPath          string `yaml:"preview_path"`
	PreviewCookieName    string `yaml:"preview_cookie_name"`
	PreviewCookieTTLHour int    `yaml:"preview_cookie_ttl_hours"`
	PreviewCookieSecure  bool   `yaml:"preview_cookie_secure"`
}

type Site struct {
	ComingSoon ComingSoon `yaml:"coming_soon"`
}

type Config struct {
	Server       Server   `yaml:"server"`
	Database     Database `yaml:"database"`
	Auth         Auth     `yaml:"auth"`
	Email        Email    `yaml:"email"`
	Site         Site     `yaml:"site"`
	InternalAuth struct {
		Token string `yaml:"token"`
	} `yaml:"internal_auth"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if cfg.Site.ComingSoon.PreviewCookieName == "" {
		cfg.Site.ComingSoon.PreviewCookieName = "ap_preview_access"
	}

	if cfg.Site.ComingSoon.PreviewCookieTTLHour <= 0 {
		cfg.Site.ComingSoon.PreviewCookieTTLHour = 48
	}

	cfg.Database.Password = sharedconfig.ResolveSecret("PARTNERS_CORE_DB_PASSWORD", cfg.Database.Password)
	cfg.Email.ResendAPIKey = sharedconfig.ResolveSecret("PARTNERS_RESEND_API_KEY", cfg.Email.ResendAPIKey)
	return &cfg, nil
}
