-- 009_create_patients.sql
-- Data pasien

CREATE TABLE IF NOT EXISTS patients (
    id SERIAL PRIMARY KEY,
    medical_record_number VARCHAR(20) NOT NULL UNIQUE,
    full_name VARCHAR(100) NOT NULL,
    nik VARCHAR(16),
    gender VARCHAR(10) NOT NULL CHECK (gender IN ('LAKI_LAKI', 'PEREMPUAN')),
    date_of_birth DATE NOT NULL,
    blood_type VARCHAR(5),
    phone VARCHAR(20),
    email VARCHAR(100),
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(10),
    emergency_contact_name VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    emergency_contact_relation VARCHAR(50),
    insurance_name VARCHAR(100),
    insurance_number VARCHAR(50),
    insurance_expiry DATE,
    allergies TEXT,
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_patients_mrn ON patients(medical_record_number);
CREATE INDEX idx_patients_name ON patients(full_name);
CREATE INDEX idx_patients_nik ON patients(nik);
CREATE INDEX idx_patients_phone ON patients(phone);
CREATE INDEX idx_patients_is_active ON patients(is_active);

CREATE OR REPLACE FUNCTION update_patients_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_patients_updated_at
    BEFORE UPDATE ON patients
    FOR EACH ROW
    EXECUTE FUNCTION update_patients_updated_at();

-- Sequence untuk nomor rekam medis
CREATE SEQUENCE IF NOT EXISTS mrn_seq START 1;
