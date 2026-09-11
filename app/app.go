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

const bootstrapMigration024 = `
ALTER TABLE queues DROP CONSTRAINT IF EXISTS queues_status_check;

ALTER TABLE queues ADD CONSTRAINT queues_status_check
    CHECK (status IN ('MENUNGGU', 'DIPANGGIL', 'SEDANG_DIPERIKSA', 'MENUNGGU_BAYAR', 'SELESAI', 'DIBATALKAN'));
`

const bootstrapInsuranceProviders = `
CREATE TABLE IF NOT EXISTS insurance_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO insurance_providers (name, sort_order) VALUES
    ('BPJS', 1),
    ('Umum', 2),
    ('Asuransi Lain', 3)
ON CONFLICT (name) DO NOTHING;
`

const bootstrapAppointments = `
CREATE TABLE IF NOT EXISTS appointments (
    id SERIAL PRIMARY KEY,
    appointment_number VARCHAR(20) NOT NULL UNIQUE,
    patient_id INTEGER NOT NULL REFERENCES patients(id),
    doctor_id INTEGER NOT NULL REFERENCES doctors(id),
    appointment_date DATE NOT NULL,
    start_time VARCHAR(5) NOT NULL DEFAULT '09:00',
    purpose TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'TERJADWAL',
    notes TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_appointments_date ON appointments(appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointments(status);
`

const bootstrapPainAssessments = `
CREATE TABLE IF NOT EXISTS pain_assessments (
    id SERIAL PRIMARY KEY,
    medical_record_id INTEGER NOT NULL UNIQUE REFERENCES medical_records(id),
    location TEXT,
    quality VARCHAR(20),
    intensity INTEGER NOT NULL DEFAULT 0 CHECK (intensity BETWEEN 0 AND 10),
    onset_duration TEXT,
    radiation TEXT,
    aggravating_factors TEXT,
    relieving_factors TEXT,
    notes TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pain_assessments_mr ON pain_assessments(medical_record_id);
`

const bootstrapImmunizations = `
CREATE TABLE IF NOT EXISTS immunization_schedules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    age_label VARCHAR(50) NOT NULL,
    month_from INT NOT NULL DEFAULT 0,
    month_to INT,
    category VARCHAR(20) NOT NULL DEFAULT 'WAJIB',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS patient_immunizations (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER NOT NULL REFERENCES patients(id),
    immunization_id INTEGER REFERENCES immunization_schedules(id),
    vaccination_date DATE NOT NULL,
    vaccine_name VARCHAR(100) NOT NULL,
    batch_number VARCHAR(50),
    provider_name VARCHAR(100),
    notes TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patient_imm_patient ON patient_immunizations(patient_id);

INSERT INTO immunization_schedules (name, age_label, month_from, month_to, category, sort_order) VALUES
    ('Hepatitis B 1', 'Baru lahir (0-24 jam)', 0, 0, 'WAJIB', 1),
    ('BCG', '1 bulan', 1, 1, 'WAJIB', 2),
    ('Polio 1 (OPV)', '1 bulan', 1, 1, 'WAJIB', 3),
    ('DTP-HB-Hib 1', '2 bulan', 2, 2, 'WAJIB', 4),
    ('Polio 2 (OPV)', '2 bulan', 2, 2, 'WAJIB', 5),
    ('PCV 1', '2 bulan', 2, 2, 'WAJIB', 6),
    ('Rotavirus 1', '2 bulan', 2, 2, 'WAJIB', 7),
    ('DTP-HB-Hib 2', '3 bulan', 3, 3, 'WAJIB', 8),
    ('Polio 3 (OPV)', '3 bulan', 3, 3, 'WAJIB', 9),
    ('Rotavirus 2', '3 bulan', 3, 3, 'WAJIB', 10),
    ('DTP-HB-Hib 3', '4 bulan', 4, 4, 'WAJIB', 11),
    ('Polio 4 (IPV)', '4 bulan', 4, 4, 'WAJIB', 12),
    ('PCV 2', '4 bulan', 4, 4, 'WAJIB', 13),
    ('PCV 3', '12 bulan', 12, 12, 'WAJIB', 14),
    ('Influenza 1', '6 bulan', 6, 6, 'ANJURAN', 15),
    ('MR 1', '9 bulan', 9, 9, 'WAJIB', 16),
    ('Influenza 2', '7 bulan', 7, 7, 'ANJURAN', 17),
    ('MMR', '15 bulan', 15, 15, 'WAJIB', 18),
    ('DTP-HB-Hib 4', '18 bulan', 18, 18, 'WAJIB', 19),
    ('Varicella 1', '12 bulan', 12, 12, 'ANJURAN', 20),
    ('JE (daerah endemis)', '9 bulan', 9, 9, 'ANJURAN', 21),
    ('Typhoid', '2 tahun (ulang tiap 3 th)', 24, 24, 'ANJURAN', 22),
    ('Hepatitis A', '2 tahun', 24, 24, 'ANJURAN', 23),
    ('MMR 2', '5-7 tahun', 60, 84, 'WAJIB', 24),
    ('DTP (booster) + Polio', '5-7 tahun', 60, 84, 'WAJIB', 25),
    ('HPV', '9-14 tahun', 108, 168, 'ANJURAN', 26)
ON CONFLICT (name) DO NOTHING;
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
		migrator.RunBootstrapSQL(
			bootstrapMigration022,
			bootstrapMigration023,
			bootstrapMigration024,
			bootstrapInsuranceProviders,
			bootstrapAppointments,
			bootstrapPainAssessments,
			bootstrapImmunizations,
		)
	}

	srv := server.New(cfg, db)
	handler = srv.Handler()
}

func Handler() (http.Handler, error) {
	initOnce.Do(initHandler)
	return handler, initErr
}
