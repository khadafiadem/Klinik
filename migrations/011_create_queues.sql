-- 011_create_queues.sql
-- Antrian pasien

CREATE TABLE IF NOT EXISTS queues (
    id SERIAL PRIMARY KEY,
    queue_number VARCHAR(20) NOT NULL,
    registration_id INTEGER NOT NULL REFERENCES registrations(id) ON DELETE RESTRICT,
    patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE RESTRICT,
    doctor_id INTEGER NOT NULL REFERENCES doctors(id) ON DELETE RESTRICT,
    queue_date DATE NOT NULL DEFAULT CURRENT_DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'MENUNGGU' CHECK (status IN ('MENUNGGU', 'DIPANGGIL', 'SEDANG_DIPERIKSA', 'SELESAI', 'DIBATALKAN')),
    called_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_queues_number ON queues(queue_number);
CREATE INDEX idx_queues_registration ON queues(registration_id);
CREATE INDEX idx_queues_doctor ON queues(doctor_id);
CREATE INDEX idx_queues_date ON queues(queue_date);
CREATE INDEX idx_queues_status ON queues(status);

CREATE OR REPLACE FUNCTION update_queues_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_queues_updated_at
    BEFORE UPDATE ON queues
    FOR EACH ROW
    EXECUTE FUNCTION update_queues_updated_at();

-- Sequence untuk nomor antrian per hari
CREATE SEQUENCE IF NOT EXISTS queue_daily_seq START 1;
