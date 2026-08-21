-- 014_create_treatments.sql
-- Perawatan / Tindakan

CREATE TABLE IF NOT EXISTS treatments (
    id SERIAL PRIMARY KEY,
    treatment_code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    default_cost NUMERIC(15,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_treatments_code ON treatments(treatment_code);
CREATE INDEX idx_treatments_name ON treatments(name);

-- Junction table: rekam medis <-> perawatan
CREATE TABLE IF NOT EXISTS medical_record_treatments (
    id SERIAL PRIMARY KEY,
    medical_record_id INTEGER NOT NULL REFERENCES medical_records(id) ON DELETE CASCADE,
    treatment_id INTEGER NOT NULL REFERENCES treatments(id) ON DELETE RESTRICT,
    cost NUMERIC(15,2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mrt_medical_record ON medical_record_treatments(medical_record_id);
CREATE INDEX idx_mrt_treatment ON medical_record_treatments(treatment_id);
