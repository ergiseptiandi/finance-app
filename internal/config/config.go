package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server  ServerConfig
	MySQL   MySQLConfig
	Auth    AuthConfig
	Seed    SeedConfig
	SMTP    SMTPConfig
	Storage StorageConfig
}

type ServerConfig struct {
	Port string
}

type MySQLConfig struct {
	Enabled         bool
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AuthConfig struct {
	JWTSecret       string
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type SeedConfig struct {
	UserName     string
	UserEmail    string
	UserPassword string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type StorageConfig struct {
	UploadDir string
}

func Load() (Config, error) {
	mysqlEnabled, err := getEnvBool("DB_ENABLED", false)
	if err != nil {
		return Config{}, err
	}

	maxOpenConns, err := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return Config{}, err
	}

	maxIdleConns, err := getEnvInt("DB_MAX_IDLE_CONNS", 25)
	if err != nil {
		return Config{}, err
	}

	connMaxLifetime, err := getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	accessTokenTTL, err := getEnvDuration("AUTH_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	refreshTokenTTL, err := getEnvDuration("AUTH_REFRESH_TOKEN_TTL", 24*time.Hour*30)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		MySQL: MySQLConfig{
			Enabled:         mysqlEnabled,
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnv("DB_PORT", "3306"),
			User:            os.Getenv("DB_USER"),
			Password:        os.Getenv("DB_PASSWORD"),
			Name:            os.Getenv("DB_NAME"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
		},
		Auth: AuthConfig{
			JWTSecret:       os.Getenv("AUTH_JWT_SECRET"),
			Issuer:          getEnv("AUTH_ISSUER", "finance-backend"),
			AccessTokenTTL:  accessTokenTTL,
			RefreshTokenTTL: refreshTokenTTL,
		},
		Seed: SeedConfig{
			UserName:     getEnv("SEED_USER_NAME", getEnv("AUTH_BOOTSTRAP_NAME", "Owner")),
			UserEmail:    firstNonEmptyEnv("SEED_USER_EMAIL", "AUTH_BOOTSTRAP_EMAIL"),
			UserPassword: firstNonEmptyEnv("SEED_USER_PASSWORD", "AUTH_BOOTSTRAP_PASSWORD"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.mailtrap.io"),
			Port:     func() int { p, _ := getEnvInt("SMTP_PORT", 2525); return p }(),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     getEnv("SMTP_FROM", "noreply@finance-app.local"),
		},
		Storage: StorageConfig{
			UploadDir: getEnv("UPLOAD_DIR", "uploads"),
		},
	}

	return cfg, nil
}

func (c MySQLConfig) Validate() error {
	if !c.Enabled {
		return fmt.Errorf("DB_ENABLED must be true")
	}

	if c.User == "" {
		return fmt.Errorf("DB_USER is required when DB_ENABLED=true")
	}

	if c.Name == "" {
		return fmt.Errorf("DB_NAME is required when DB_ENABLED=true")
	}

	return nil
}

func (c AuthConfig) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("AUTH_JWT_SECRET is required")
	}

	return nil
}

func (c SeedConfig) Validate() error {
	if c.UserEmail == "" {
		return fmt.Errorf("SEED_USER_EMAIL is required")
	}

	if c.UserPassword == "" {
		return fmt.Errorf("SEED_USER_PASSWORD is required")
	}

	return nil
}

func (c Config) ValidateForAPI() error {
	if err := c.MySQL.Validate(); err != nil {
		return err
	}

	return c.Auth.Validate()
}

func (c Config) ValidateForMigrate() error {
	return c.MySQL.Validate()
}

func (c Config) ValidateForSeed() error {
	if err := c.MySQL.Validate(); err != nil {
		return err
	}

	return c.Seed.Validate()
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvBool(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}

	return parsed, nil
}

func getEnvInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", key, err)
	}

	return parsed, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", key, err)
	}

	return parsed, nil
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}

	return ""
}
