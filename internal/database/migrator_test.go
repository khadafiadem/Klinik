package database

import (
	"os"
	"path/filepath"
	"testing"

	"klinik-app/internal/logger"
)

func TestMain(m *testing.M) {
	logger.Init("debug")
	os.Exit(m.Run())
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"001_create_users.sql", "001"},
		{"002_create_roles.sql", "002"},
		{"010_add_index.sql", "010"},
		{"invalid.sql", ""},
		{"no_underscore.sql", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := extractVersion(tt.filename)
			if result != tt.expected {
				t.Errorf("extractVersion(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestLoadMigrationFilesEmpty(t *testing.T) {
	dir := t.TempDir()

	m := &Migrator{migrationsDir: dir}
	migrations, err := m.loadMigrationFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}

func TestLoadMigrationFilesNonExistentDir(t *testing.T) {
	m := &Migrator{migrationsDir: filepath.Join(t.TempDir(), "nonexistent")}
	migrations, err := m.loadMigrationFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 0 {
		t.Errorf("expected 0 migrations, got %d", len(migrations))
	}
}

func TestLoadMigrationFiles(t *testing.T) {
	dir := t.TempDir()

	content := "CREATE TABLE test (id SERIAL PRIMARY KEY);"
	err := os.WriteFile(filepath.Join(dir, "001_create_test.sql"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test migration: %v", err)
	}

	m := &Migrator{migrationsDir: dir}
	migrations, err := m.loadMigrationFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(migrations) != 1 {
		t.Fatalf("expected 1 migration, got %d", len(migrations))
	}
	if migrations[0].Version != "001" {
		t.Errorf("expected version '001', got '%s'", migrations[0].Version)
	}
	if migrations[0].Filename != "001_create_test.sql" {
		t.Errorf("expected filename '001_create_test.sql', got '%s'", migrations[0].Filename)
	}
}
