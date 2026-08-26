package app

import (
	"net/http"
	"sync"

	"klinik-app/internal/config"
	"klinik-app/internal/database"
	"klinik-app/internal/logger"
	"klinik-app/internal/server"
)

var (
	initOnce sync.Once
	handler  http.Handler
	initErr  error
)

const bootstrapMigration022 = `
CREATE SEQUENCE IF NOT EXISTS kiosk_queue_seq START 1;

DO $$ BEGIN
    ALTER TABLE queues ADD COLUMN IF NOT EXISTS queue_source VARCHAR(20) NOT NULL DEFAULT 'ADMIN';
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE queues ADD COLUMN IF NOT EXISTS called_by INTEGER REFERENCES users(id);
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE queues ADD COLUMN IF NOT EXISTS doctor_name_snapshot VARCHAR(100);
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE queues ALTER COLUMN registration_id DROP NOT NULL;
EXCEPTION WHEN others THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE queues ALTER COLUMN patient_id DROP NOT NULL;
EXCEPTION WHEN others THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE queues ALTER COLUMN doctor_id DROP NOT NULL;
EXCEPTION WHEN others THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE queues ADD CONSTRAINT queues_status_check
        CHECK (status IN ('MENUNGGU', 'DIPANGGIL', 'SEDANG_DIPERIKSA', 'SELESAI', 'DIBATALKAN'));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_queues_source ON queues(queue_source);
`

const bootstrapMigration023 = `
CREATE TABLE IF NOT EXISTS queue_config (
    id INT PRIMARY KEY DEFAULT 1,
    paused BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
INSERT INTO queue_config (id, paused) VALUES (1, FALSE)
    ON CONFLICT (id) DO NOTHING;
`

func initHandler() {
	logger.Init("info")

	cfg, err := config.Load()
	if err != nil {
		initErr = err
		return
	}
	logger.Init(cfg.LogLevel)

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error.Printf("Database connection failed: %v", err)
		db = nil
	}

	if db != nil {
		migrator := database.NewMigrator(db, "migrations")
		if err := migrator.Up(); err != nil {
			logger.Info.Printf("File migration skipped (Vercel): %v", err)
		}
		migrator.RunBootstrapSQL(bootstrapMigration022)
		migrator.RunBootstrapSQL(bootstrapMigration023)
	}

	srv := server.New(cfg, db)
	handler = srv.Handler()
}

func Handler() (http.Handler, error) {
	initOnce.Do(initHandler)
	return handler, initErr
}
