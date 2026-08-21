package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"klinik-app/internal/logger"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info.Println("Database connected successfully")
	return db, nil
}

func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
