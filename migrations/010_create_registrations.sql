-- 010_create_registrations.sql
-- Pendaftaran pasien

CREATE TABLE IF NOT EXISTS registrations (
    id SERIAL PRIMARY KEY,
    registration_number VARCHAR(20) NOT NULL UNIQUE,
    patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id INTEGER NOT NULL REFERENCES doctors(id) ON DELETE RESTRICT,
    registration_date DATE NOT NULL DEFAULT CURRENT_DATE,
    registration_type VARCHAR(20) NOT NULL DEFAULT 'UMUM' CHECK (registration_type IN ('UMUM', 'BPJS', 'ASURANSI')),
    complaint TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'TERDAFTAR' CHECK (status IN ('TERDAFTAR', 'SEDANG_DIPERIKSA', 'SELESAI', 'DIBATALKAN')),
    notes TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_registrations_number ON registrations(registration_number);
CREATE INDEX idx_registrations_patient ON registrations(patient_id);
CREATE INDEX idx_registrations_doctor ON registrations(doctor_id);
CREATE INDEX idx_registrations_date ON registrations(registration_date);
CREATE INDEX idx_registrations_status ON registrations(status);

CREATE OR REPLACE FUNCTION update_registrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_registrations_updated_at
    BEFORE UPDATE ON registrations
    FOR EACH ROW
    EXECUTE FUNCTION update_registrations_updated_at();

-- Sequence untuk nomor pendaftaran
CREATE SEQUENCE IF NOT EXISTS reg_seq START 1;
