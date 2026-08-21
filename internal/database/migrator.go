package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"klinik-app/internal/logger"
)

type Migration struct {
	Version  string
	Filename string
	SQL      string
}

type Migrator struct {
	db            *sql.DB
	migrationsDir string
}

func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

func (m *Migrator) ensureMigrationsTable() error {
	query := `CREATE TABLE IF NOT EXISTS schema_migrations (
		id SERIAL PRIMARY KEY,
		version VARCHAR(255) NOT NULL UNIQUE,
		filename VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	return nil
}

func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, nil
}

func (m *Migrator) loadMigrationFiles() ([]Migration, error) {
	entries, err := os.ReadDir(m.migrationsDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info.Println("Migrations directory does not exist, skipping")
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := extractVersion(entry.Name())
		if version == "" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(m.migrationsDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version:  version,
			Filename: entry.Name(),
			SQL:      string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) Up() error {
	if err := m.ensureMigrationsTable(); err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	migrations, err := m.loadMigrationFiles()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		logger.Info.Println("No migrations to apply")
		return nil
	}

	count := 0
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		logger.Info.Printf("Applying migration: %s (%s)", migration.Version, migration.Filename)

		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migration.Version, err)
		}

		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, filename) VALUES ($1, $2)",
			migration.Version, migration.Filename,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
		}

		count++
		logger.Info.Printf("Migration %s applied successfully", migration.Version)
	}

	logger.Info.Printf("Applied %d migration(s)", count)
	return nil
}

type MigrationStatus struct {
	Version  string
	Filename string
	Applied  bool
}

func (m *Migrator) Status() ([]MigrationStatus, error) {
	if err := m.ensureMigrationsTable(); err != nil {
		return nil, err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return nil, err
	}

	migrations, err := m.loadMigrationFiles()
	if err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, migration := range migrations {
		statuses = append(statuses, MigrationStatus{
			Version:  migration.Version,
			Filename: migration.Filename,
			Applied:  applied[migration.Version],
		})
	}

	return statuses, nil
}

func extractVersion(filename string) string {
	name := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(name, "_", 2)
	if len(parts) < 2 {
		return ""
	}
	version := parts[0]
	for _, ch := range version {
		if ch < '0' || ch > '9' {
			return ""
		}
	}
	if len(version) == 0 {
		return ""
	}
	return version
}
