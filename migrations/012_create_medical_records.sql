-- 012_create_medical_records.sql
-- Rekam Medis

CREATE TABLE IF NOT EXISTS medical_records (
    id SERIAL PRIMARY KEY,
    medical_record_number VARCHAR(30) NOT NULL UNIQUE,
    patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id INTEGER NOT NULL REFERENCES doctors(id) ON DELETE RESTRICT,
    registration_id INTEGER REFERENCES registrations(id) ON DELETE SET NULL,
    examination_date DATE NOT NULL DEFAULT CURRENT_DATE,
    chief_complaint TEXT,
    vital_signs JSONB,
    anamnesis TEXT,
    physical_examination TEXT,
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'FINAL', 'AMENDED')),
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_medical_records_patient ON medical_records(patient_id);
CREATE INDEX idx_medical_records_doctor ON medical_records(doctor_id);
CREATE INDEX idx_medical_records_registration ON medical_records(registration_id);
CREATE INDEX idx_medical_records_date ON medical_records(examination_date);
CREATE INDEX idx_medical_records_number ON medical_records(medical_record_number);

CREATE OR REPLACE FUNCTION update_medical_records_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_medical_records_updated_at
    BEFORE UPDATE ON medical_records
    FOR EACH ROW
    EXECUTE FUNCTION update_medical_records_updated_at();

CREATE SEQUENCE IF NOT EXISTS mr_seq START 1;
