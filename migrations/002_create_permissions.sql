-- 002_create_permissions.sql
-- Tabel permissions untuk granular access control

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk pencarian permission berdasarkan nama
CREATE INDEX idx_permissions_name ON permissions(name);

-- Insert default permissions
INSERT INTO permissions (name, description) VALUES
    ('patients.read', 'Melihat data pasien'),
    ('patients.write', 'Menambah/edit data pasien'),
    ('doctors.read', 'Melihat data dokter'),
    ('doctors.write', 'Menambah/edit data dokter'),
    ('registrations.read', 'Melihat pendaftaran'),
    ('registrations.write', 'Menambah/edit pendaftaran'),
    ('queues.read', 'Melihat antrian'),
    ('queues.write', 'Mengubah status antrian'),
    ('medical_records.read', 'Melihat rekam medis'),
    ('medical_records.write', 'Menambah/edit rekam medis'),
    ('prescriptions.read', 'Melihat resep'),
    ('prescriptions.write', 'Menambah/edit resep'),
    ('medicines.read', 'Melihat data obat'),
    ('medicines.write', 'Menambah/edit data obat'),
    ('pharmacy.read', 'Melihat transaksi apotek'),
    ('pharmacy.write', 'Memproses transaksi apotek'),
    ('payments.read', 'Melihat pembayaran'),
    ('payments.write', 'Memproses pembayaran'),
    ('reports.read', 'Melihat laporan'),
    ('users.read', 'Melihat data pengguna'),
    ('users.write', 'Menambah/edit pengguna'),
    ('settings.read', 'Melihat pengaturan'),
    ('settings.write', 'Mengubah pengaturan'),
    ('audit.read', 'Melihat log audit')
ON CONFLICT (name) DO NOTHING;
