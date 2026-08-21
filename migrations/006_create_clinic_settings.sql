-- 006_create_clinic_settings.sql
-- Pengaturan klinik

CREATE TABLE IF NOT EXISTS clinic_settings (
    id SERIAL PRIMARY KEY,
    clinic_name VARCHAR(200) NOT NULL DEFAULT 'Klinik',
    clinic_address TEXT,
    clinic_phone VARCHAR(20),
    clinic_email VARCHAR(100),
    clinic_logo VARCHAR(255),
    opening_time TIME DEFAULT '08:00:00',
    closing_time TIME DEFAULT '21:00:00',
    max_patients_per_day INTEGER DEFAULT 100,
    registration_fee DECIMAL(10,2) DEFAULT 0,
    consultation_fee DECIMAL(10,2) DEFAULT 0,
    tax_percentage DECIMAL(5,2) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'IDR',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default settings
INSERT INTO clinic_settings (clinic_name) VALUES ('Klinik') ON CONFLICT DO NOTHING;

CREATE OR REPLACE FUNCTION update_clinic_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_clinic_settings_updated_at
    BEFORE UPDATE ON clinic_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_clinic_settings_updated_at();
