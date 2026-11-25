package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration loaded from environment variables.
type Config struct {
	// Database connection string (DSN).
	DatabaseURL string

	// HTTP server configuration.
	HTTPAddr string
	// Maximum duration for reading requests.
	HTTPReadTimeout time.Duration
	// Maximum duration for writing responses.
	HTTPWriteTimeout time.Duration
	// Maximum duration for idle connections.
	HTTPIdleTimeout time.Duration

	// Public base URL for generating absolute URLs.
	PublicBaseURL string

	// Photo storage directory on the filesystem.
	PhotoStorageDir string

	// Session configuration.
	SessionCookieSecure bool
	SessionTTL          time.Duration

	// Invitation configuration.
	InvitationDefaultTTL time.Duration

	// Maximum upload size in bytes (default 12 MiB).
	MaxUploadBytes int64

	// Log level: debug, info, warn, error.
	LogLevel string

	// Trusted proxies for rate limiting and CORS.
	TrustedProxies []string
}

// Load reads configuration from environment variables.
// All required fields must be present; otherwise an error is returned.
func Load() (*Config, error) {
	c := &Config{
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		PublicBaseURL:      getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		PhotoStorageDir:    getEnv("PHOTO_STORAGE_DIR", "./photos"),
		SessionCookieSecure: parseBool("SESSION_COOKIE_SECURE", false),
		SessionTTL:         parseDuration("SESSION_TTL", 24*time.Hour),
		InvitationDefaultTTL: parseDuration("INVITATION_DEFAULT_TTL", 7*24*time.Hour),
		MaxUploadBytes:     parseInt64("MAX_UPLOAD_BYTES", 12*1024*1024),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

func parseInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return fallback
}
