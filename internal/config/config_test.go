package config

import (
	"os"
	"testing"
)

func TestLoadSuccess(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "secret")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("APP_ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Errorf("expected AppEnv 'test', got '%s'", cfg.AppEnv)
	}
	if cfg.AppPort != 9090 {
		t.Errorf("expected AppPort 9090, got %d", cfg.AppPort)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("unexpected DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "secret" {
		t.Errorf("unexpected JWTSecret: %s", cfg.JWTSecret)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got '%s'", cfg.LogLevel)
	}
}

func TestLoadDefaults(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "secret")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("APP_PORT")
		os.Unsetenv("APP_ENV")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppPort != 8080 {
		t.Errorf("expected default AppPort 8080, got %d", cfg.AppPort)
	}
	if cfg.AppEnv != "development" {
		t.Errorf("expected default AppEnv 'development', got '%s'", cfg.AppEnv)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default LogLevel 'info', got '%s'", cfg.LogLevel)
	}
}

func TestLoadMissingDatabaseURL(t *testing.T) {
	os.Setenv("JWT_SECRET", "secret")
	os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("JWT_SECRET")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestLoadMissingJWTSecret(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET")
	}
}

func TestLoadInvalidPort(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("JWT_SECRET", "secret")
	os.Setenv("APP_PORT", "notanumber")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("APP_PORT")
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid APP_PORT")
	}
}
