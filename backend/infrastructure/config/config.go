package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	Auth     AuthConfig
}

type AppConfig struct {
	Name        string
	Env         string
	Version     string
	Port        int
	WindowTitle string
	Width       int
	Height      int
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpen      int
	MaxIdle      int
	MaxLifetime  time.Duration
	MigrationDir string
}

type LoggerConfig struct {
	Level  string
	Format string
	Output string
}

type AuthConfig struct {
	ArgonMemory      uint32
	ArgonIterations  uint32
	ArgonParallelism uint8
	ArgonSaltLength  uint32
	ArgonKeyLength   uint32
	SessionTTL       time.Duration
	LockoutTTL       time.Duration
	MaxLoginAttempts int
}

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "vfinancy"),
			Env:         getEnv("APP_ENV", "development"),
			Version:     getEnv("APP_VERSION", "0.0.0"),
			Port:        getEnvInt("APP_PORT", 0),
			WindowTitle: getEnv("APP_WINDOW_TITLE", "vfinancy"),
			Width:       getEnvInt("APP_WIDTH", 1280),
			Height:      getEnvInt("APP_HEIGHT", 800),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "casa123"),
			Name:         getEnv("DB_NAME", "vfinancy"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpen:      getEnvInt("DB_MAX_OPEN", 25),
			MaxIdle:      getEnvInt("DB_MAX_IDLE", 5),
			MaxLifetime:  time.Duration(getEnvInt("DB_MAX_LIFETIME_MIN", 30)) * time.Minute,
			MigrationDir: getEnv("DB_MIGRATION_DIR", "migrations"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Auth: AuthConfig{
			ArgonMemory:      uint32(getEnvInt("AUTH_ARGON_MEMORY_KB", 65536)),
			ArgonIterations:  uint32(getEnvInt("AUTH_ARGON_ITERATIONS", 3)),
			ArgonParallelism: uint8(getEnvInt("AUTH_ARGON_PARALLELISM", 2)),
			ArgonSaltLength:  uint32(getEnvInt("AUTH_ARGON_SALT_LEN", 16)),
			ArgonKeyLength:   uint32(getEnvInt("AUTH_ARGON_KEY_LEN", 32)),
			SessionTTL:       time.Duration(getEnvInt("AUTH_SESSION_TTL_MIN", 60)) * time.Minute,
			LockoutTTL:       time.Duration(getEnvInt("AUTH_LOCKOUT_TTL_MIN", 15)) * time.Minute,
			MaxLoginAttempts: getEnvInt("AUTH_MAX_LOGIN_ATTEMPTS", 5),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("config: DB_HOST is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("config: DB_NAME is required")
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}

func (c *DatabaseConfig) AdminDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
