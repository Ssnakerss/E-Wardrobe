package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad_ValidEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "postgres://test:test@localhost/test", cfg.DatabaseURL)
	require.Equal(t, ":8080", cfg.HTTPAddr)
	require.Equal(t, "http://localhost:8080", cfg.PublicBaseURL)
	require.Equal(t, "./photos", cfg.PhotoStorageDir)
	require.False(t, cfg.SessionCookieSecure)
	require.Equal(t, 24*time.Hour, cfg.SessionTTL)
	require.Equal(t, 7*24*time.Hour, cfg.InvitationDefaultTTL)
	require.Equal(t, int64(12*1024*1024), cfg.MaxUploadBytes)
	require.Equal(t, "info", cfg.LogLevel)
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://custom:pass@host/db")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("PUBLIC_BASE_URL", "https://example.com")
	t.Setenv("PHOTO_STORAGE_DIR", "/data/photos")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	t.Setenv("SESSION_TTL", "2h")
	t.Setenv("INVITATION_DEFAULT_TTL", "72h")
	t.Setenv("MAX_UPLOAD_BYTES", "5242880")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "postgres://custom:pass@host/db", cfg.DatabaseURL)
	require.Equal(t, ":9090", cfg.HTTPAddr)
	require.Equal(t, "https://example.com", cfg.PublicBaseURL)
	require.Equal(t, "/data/photos", cfg.PhotoStorageDir)
	require.True(t, cfg.SessionCookieSecure)
	require.Equal(t, 2*time.Hour, cfg.SessionTTL)
	require.Equal(t, 3*24*time.Hour, cfg.InvitationDefaultTTL)
	require.Equal(t, int64(5242880), cfg.MaxUploadBytes)
	require.Equal(t, "debug", cfg.LogLevel)
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	// Make sure DATABASE_URL is not set
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestLoad_InvalidSessionTTL_UsesDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	t.Setenv("SESSION_TTL", "not-a-duration")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, 24*time.Hour, cfg.SessionTTL)
}

func TestLoad_InvalidMaxUploadBytes_UsesDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	t.Setenv("MAX_UPLOAD_BYTES", "not-a-number")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, int64(12*1024*1024), cfg.MaxUploadBytes)
}
