-- 013_create_diagnoses.sql
-- Diagnosa

CREATE TABLE IF NOT EXISTS diagnoses (
    id SERIAL PRIMARY KEY,
    diagnosis_code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_diagnoses_code ON diagnoses(diagnosis_code);
CREATE INDEX idx_diagnoses_name ON diagnoses(name);

-- Junction table: rekam medis <-> diagnosa
CREATE TABLE IF NOT EXISTS medical_record_diagnoses (
    id SERIAL PRIMARY KEY,
    medical_record_id INTEGER NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    diagnosis_id INTEGER NOT NULL REFERENCES diagnoses(id) ON DELETE RESTRICT,
    diagnosis_type VARCHAR(20) NOT NULL DEFAULT 'UTAMA' CHECK (diagnosis_type IN ('UTAMA', 'SEKUNDER')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mrd_medical_record ON medical_record_diagnoses(medical_record_id);
CREATE INDEX idx_mrd_diagnosis ON medical_record_diagnoses(diagnosis_id);
