package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	Sync     SyncConfig
}

type AppConfig struct {
	Name        string
	Env         string
	Version     string
	WindowTitle string
	Width       int
	Height      int
}

// DatabaseConfig holds the connection settings for the local runtime
// database. The desktop app runs on SQLite (offline-first, single
// source of truth).
type DatabaseConfig struct {
	Driver       string
	Path         string
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

// SyncConfig configures the background replication to the cloud
// PostgreSQL mirror over a direct connection. When Disabled the app
// runs fully offline (SQLite remains the runtime database either way).
type SyncConfig struct {
	Enabled      bool
	Host         string
	Port         int
	Name         string
	User         string
	Password     string
	SSLMode      string
	PollInterval time.Duration
}

func Load() (*Config, error) {
	godotenv.Load() // ponytail: .env optional; explicit env vars win
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "vfinancy"),
			Env:         getEnv("APP_ENV", "development"),
			Version:     getEnv("APP_VERSION", "0.0.0"),
			WindowTitle: getEnv("APP_WINDOW_TITLE", "vfinancy"),
			Width:       getEnvInt("APP_WIDTH", 1280),
			Height:      getEnvInt("APP_HEIGHT", 800),
		},
		Database: DatabaseConfig{
			Driver:       getEnv("DB_DRIVER", "sqlite"),
			Path:         getEnv("DB_PATH", "data/vfinancy.db"),
			MaxOpen:      getEnvInt("DB_MAX_OPEN", 25),
			MaxIdle:      getEnvInt("DB_MAX_IDLE", 5),
			MaxLifetime:  time.Duration(getEnvInt("DB_MAX_LIFETIME_MIN", 30)) * time.Minute,
			MigrationDir: getEnv("DB_MIGRATION_DIR", "migrations/sqlite"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Sync: SyncConfig{
			Enabled:      getEnvBool("SYNC_ENABLED", false),
			Host:         getEnv("SYNC_DB_HOST", ""),
			Port:         getEnvInt("SYNC_DB_PORT", 5432),
			Name:         getEnv("SYNC_DB_NAME", ""),
			User:         getEnv("SYNC_DB_USER", ""),
			Password:     getEnv("SYNC_DB_PASSWORD", ""),
			SSLMode:      getEnv("SYNC_SSLMODE", "require"),
			PollInterval: time.Duration(getEnvInt("SYNC_POLL_INTERVAL_SEC", 30)) * time.Second,
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	switch c.Database.Driver {
	case "sqlite", "postgres":
	default:
		return fmt.Errorf("config: DB_DRIVER must be sqlite or postgres, got %q", c.Database.Driver)
	}
	if c.Database.Driver == "sqlite" && c.Database.Path == "" {
		return fmt.Errorf("config: DB_PATH is required when DB_DRIVER=sqlite")
	}
	if c.Sync.Enabled && (c.Sync.Host == "" || c.Sync.Name == "" || c.Sync.User == "") {
		return fmt.Errorf("config: SYNC_DB_HOST, SYNC_DB_NAME and SYNC_DB_USER are required when sync is enabled")
	}
	return nil
}

// DSN returns the PostgreSQL connection string for the cloud mirror.
func (c *SyncConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
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

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
