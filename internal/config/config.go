package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	WhatsApp WhatsAppConfig `yaml:"whatsapp"`
	Webhook  WebhookConfig  `yaml:"webhook"`
	Logging  LoggingConfig  `yaml:"logging"`

	ConfigPath string `yaml:"-"`
}

type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

type DatabaseConfig struct {
	Backend    string         `yaml:"backend"`
	PostgreSQL PostgresConfig `yaml:"postgresql"`
	SQLite     SQLiteConfig   `yaml:"sqlite"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	MaxConns int    `yaml:"max_conns"`
	MinConns int    `yaml:"min_conns"`
}

func (p PostgresConfig) DSN() string {
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     fmt.Sprintf("%s:%d", p.Host, p.Port),
		Path:     p.DBName,
		RawQuery: fmt.Sprintf("sslmode=%s", url.QueryEscape(p.SSLMode)),
	}
	return u.String()
}

type SQLiteConfig struct {
	Path string `yaml:"path"`
}

type AuthConfig struct {
	GlobalAPIKey string `yaml:"global_api_key"`
}

type WhatsAppConfig struct {
	SessionDBPath string `yaml:"session_db_path"`
}

type WebhookConfig struct {
	GlobalURL     string            `yaml:"global_url"`
	GlobalEvents  []string          `yaml:"global_events"`
	GlobalHeaders map[string]string `yaml:"global_headers"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8550,
			BaseURL: "http://localhost:8550",
		},
		Database: DatabaseConfig{
			Backend: "postgresql",
			PostgreSQL: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "whatsgo",
				SSLMode:  "disable",
				MaxConns: 50,
				MinConns: 10,
			},
			SQLite: SQLiteConfig{
				Path: "./data/whatsgo.db",
			},
		},
		WhatsApp: WhatsAppConfig{
			SessionDBPath: "./data/sessions.db",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func Load() (*Config, error) {
	cfg := Default()

	configPath := os.Getenv("WHATSGO_CONFIG")
	if configPath == "" {
		configPath = "whatsgo.yaml"
	}
	cfg.ConfigPath = configPath

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	expanded := expandEnvVars(string(data))

	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Auth.GlobalAPIKey != "" && len(c.Auth.GlobalAPIKey) < 16 {
		return fmt.Errorf("auth.global_api_key must be at least 16 characters (got %d)", len(c.Auth.GlobalAPIKey))
	}

	if c.Server.BaseURL == "" {
		return fmt.Errorf("server.base_url must not be empty")
	}
	u, err := url.Parse(c.Server.BaseURL)
	if err != nil {
		return fmt.Errorf("server.base_url is not a valid URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("server.base_url must have a scheme and host (got %q)", c.Server.BaseURL)
	}

	switch c.Database.Backend {
	case "postgresql", "postgres", "":
		// ok
	case "sqlite":
		if c.Database.SQLite.Path == "" {
			return fmt.Errorf("database.sqlite.path must not be empty when backend is sqlite")
		}
	default:
		return fmt.Errorf("unsupported database backend: %s", c.Database.Backend)
	}

	return nil
}

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")

		if idx := strings.Index(key, ":"); idx > 0 {
			envKey := key[:idx]
			defaultVal := key[idx+1:]
			// Support ${VAR:-default} format (strip optional leading -)
			defaultVal = strings.TrimPrefix(defaultVal, "-")
			if val, ok := os.LookupEnv(envKey); ok {
				return val
			}
			return defaultVal
		}

		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return match
	})
}
