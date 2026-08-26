-- 022_queue_kiosk_support.sql
-- Mendukung sistem antrian kiosk (layar sentuh) dan tampilan TV

-- 1. Buat sequence khusus untuk nomor antrian kiosk agar tidak ada duplikat
CREATE SEQUENCE IF NOT EXISTS kiosk_queue_seq START 1;

-- 2. Tambah kolom untuk kiosk dan monitoring
ALTER TABLE queues ADD COLUMN IF NOT EXISTS queue_source VARCHAR(20) NOT NULL DEFAULT 'ADMIN'
    CHECK (queue_source IN ('ADMIN', 'KIOSK'));
ALTER TABLE queues ADD COLUMN IF NOT EXISTS called_by INTEGER REFERENCES users(id);
ALTER TABLE queues ADD COLUMN IF NOT EXISTS doctor_name_snapshot VARCHAR(100);

-- 3. Buat registration_id bisa NULL untuk antrian kiosk (belum terdaftar)
ALTER TABLE queues ALTER COLUMN registration_id DROP NOT NULL;
ALTER TABLE queues ALTER COLUMN patient_id DROP NOT NULL;
ALTER TABLE QUEUEs ALTER COLUMN doctor_id DROP NOT NULL;

-- 4. Tambah status baru untuk antrian kiosk yang belum didaftarkan
-- MENUNGGU tetap dipakai, tapi registration_id/patient_id/doctor_id bisa NULL
-- Tambah CHECK constraint yang lebih longgar
ALTER TABLE queues DROP CONSTRAINT IF EXISTS queues_status_check;
ALTER TABLE queues ADD CONSTRAINT queues_status_check
    CHECK (status IN ('MENUNGGU', 'DIPANGGIL', 'SEDANG_DIPERIKSA', 'SELESAI', 'DIBATALKAN'));

-- 5. Index untuk query kiosk
CREATE INDEX IF NOT EXISTS idx_queues_source ON queues(queue_source);
CREATE INDEX IF NOT EXISTS idx_queues_pending_kiosk ON queues(queue_date, status, queue_source)
    WHERE queue_source = 'KIOSK' AND registration_id IS NULL AND status = 'MENUNGGU';

-- 6. Fungsi untuk generate nomor antrian kiosk secara atomic
CREATE OR REPLACE FUNCTION generate_kiosk_queue_number(p_date DATE)
RETURNS VARCHAR(20) AS $$
DECLARE
    next_val INTEGER;
    queue_num VARCHAR(20);
BEGIN
    next_val := nextval('kiosk_queue_seq');
    queue_num := 'A-' || LPAD(next_val::TEXT, 3, '0');
    RETURN queue_num;
END;
$$ LANGUAGE plpgsql;
