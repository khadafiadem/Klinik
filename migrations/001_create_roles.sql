-- 001_create_roles.sql
-- Tabel roles untuk role-based access control

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk pencarian role berdasarkan nama
CREATE INDEX idx_roles_name ON roles(name);

-- Insert default roles sesuai AGENTS.md
INSERT INTO roles (name, description) VALUES
    ('ADMIN', 'Akses penuh ke seluruh sistem'),
    ('DOCTOR', 'Akses ke data pasien, antrian, pemeriksaan, rekam medis, diagnosis, perawatan, resep'),
    ('NURSE', 'Akses ke pendaftaran pasien, antrian, pemeriksaan dasar, tanda vital, informasi pasien'),
    ('PHARMACIST', 'Akses ke resep, obat, stok, transaksi apotek'),
    ('CASHIER', 'Akses ke invoice, pembayaran, transaksi'),
    ('OWNER', 'Akses ke dashboard, laporan, laporan keuangan, laporan operasional')
ON CONFLICT (name) DO NOTHING;
