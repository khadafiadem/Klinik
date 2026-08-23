-- 021_bpjs_antrean.sql
-- Integrasi BPJS Kesehatan - Antrean Online (Mobile JKN)

-- Konfigurasi koneksi BPJS (single row)
CREATE TABLE IF NOT EXISTS bpjs_config (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mode VARCHAR(20) NOT NULL DEFAULT 'SANDBOX' CHECK (mode IN ('SANDBOX', 'PRODUCTION')),
    base_url TEXT NOT NULL DEFAULT 'https://apijkn-dev.bpjs-kesehatan.go.id/antreanfktp_dev',
    cons_id VARCHAR(100) NOT NULL DEFAULT '',
    secret_key TEXT NOT NULL DEFAULT '',
    user_key TEXT NOT NULL DEFAULT '',
    kode_ppk VARCHAR(20) NOT NULL DEFAULT '',
    nama_ppk VARCHAR(200) NOT NULL DEFAULT '',
    kode_poli VARCHAR(10) NOT NULL DEFAULT '',
    nama_poli VARCHAR(100) NOT NULL DEFAULT '',
    jam_praktek VARCHAR(50) NOT NULL DEFAULT '08:00-16:00',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO bpjs_config (id) VALUES (1) ON CONFLICT DO NOTHING;

-- Mapping dokter internal -> kode dokter BPJS
CREATE TABLE IF NOT EXISTS bpjs_doctor_map (
    id SERIAL PRIMARY KEY,
    doctor_id INTEGER NOT NULL UNIQUE REFERENCES doctors(id) ON DELETE CASCADE,
    bpjs_code VARCHAR(30) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Log sinkronisasi antrean ke BPJS
CREATE TABLE IF NOT EXISTS bpjs_log (
    id SERIAL PRIMARY KEY,
    queue_id INTEGER REFERENCES queues(id) ON DELETE SET NULL,
    action VARCHAR(30) NOT NULL,
    request_payload JSONB,
    response_code VARCHAR(10),
    response_message TEXT,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    is_sandbox BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bpjs_log_queue ON bpjs_log(queue_id);
CREATE INDEX idx_bpjs_log_created ON bpjs_log(created_at);
