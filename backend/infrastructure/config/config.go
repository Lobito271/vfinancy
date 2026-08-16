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
	Auth     AuthConfig
	Sync     SyncConfig
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

// Driver identifies the primary runtime database engine. The desktop
// app runs on SQLite by default ("sqlite"); "postgres" is supported
// for the cloud mirror and development.
type DatabaseConfig struct {
	Driver       string
	Path         string
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

// SyncConfig configures the background synchronizer that pushes local
// writes to the cloud PostgreSQL mirror and pulls remote changes back.
// When Disabled the app runs fully offline (SQLite remains the runtime
// database either way).
type SyncConfig struct {
	Enabled      bool
	ServerURL    string
	APIKey       string
	PollInterval time.Duration
}

func Load() (*Config, error) {
	godotenv.Load() // ponytail: .env optional; explicit env vars win
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
			Driver:       getEnv("DB_DRIVER", "sqlite"),
			Path:         getEnv("DB_PATH", "data/vfinancy.db"),
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "casa123"),
			Name:         getEnv("DB_NAME", "vfinancy"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
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
		Sync: SyncConfig{
			Enabled:      getEnvBool("SYNC_ENABLED", false),
			ServerURL:    getEnv("SYNC_SERVER_URL", ""),
			APIKey:       getEnv("SYNC_API_KEY", ""),
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
	if c.Database.Driver == "postgres" {
		if c.Database.Host == "" {
			return fmt.Errorf("config: DB_HOST is required")
		}
		if c.Database.Name == "" {
			return fmt.Errorf("config: DB_NAME is required")
		}
	}
	if c.Sync.Enabled && c.Sync.ServerURL == "" {
		return fmt.Errorf("config: SYNC_SERVER_URL is required when sync is enabled")
	}
	return nil
}

// ListenAddr returns the address the sync server should bind to, using
// APP_PORT (default 8787).
func (c *Config) ListenAddr() string {
	port := c.App.Port
	if port == 0 {
		port = 8787
	}
	return fmt.Sprintf(":%d", port)
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

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
