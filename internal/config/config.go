package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Auth   AuthConfig
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
	BootstrapName     string
	BootstrapEmail    string
	BootstrapPassword string
	JWTSecret         string
	Issuer            string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
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
			BootstrapName:     getEnv("AUTH_BOOTSTRAP_NAME", "Owner"),
			BootstrapEmail:    os.Getenv("AUTH_BOOTSTRAP_EMAIL"),
			BootstrapPassword: os.Getenv("AUTH_BOOTSTRAP_PASSWORD"),
			JWTSecret:         os.Getenv("AUTH_JWT_SECRET"),
			Issuer:            getEnv("AUTH_ISSUER", "finance-backend"),
			AccessTokenTTL:    accessTokenTTL,
			RefreshTokenTTL:   refreshTokenTTL,
		},
	}

	if err := cfg.MySQL.Validate(); err != nil {
		return Config{}, err
	}

	if err := cfg.Auth.Validate(cfg.MySQL.Enabled); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c MySQLConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.User == "" {
		return fmt.Errorf("DB_USER is required when DB_ENABLED=true")
	}

	if c.Name == "" {
		return fmt.Errorf("DB_NAME is required when DB_ENABLED=true")
	}

	return nil
}

func (c AuthConfig) Validate(mysqlEnabled bool) error {
	if !mysqlEnabled {
		return nil
	}

	if c.BootstrapEmail == "" {
		return fmt.Errorf("AUTH_BOOTSTRAP_EMAIL is required when DB_ENABLED=true")
	}

	if c.BootstrapPassword == "" {
		return fmt.Errorf("AUTH_BOOTSTRAP_PASSWORD is required when DB_ENABLED=true")
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("AUTH_JWT_SECRET is required when DB_ENABLED=true")
	}

	return nil
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
