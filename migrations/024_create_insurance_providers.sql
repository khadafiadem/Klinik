-- 024_create_insurance_providers.sql
CREATE TABLE IF NOT EXISTS insurance_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO insurance_providers (name, sort_order) VALUES
    ('BPJS', 1),
    ('Umum', 2),
    ('Asuransi Lain', 3)
ON CONFLICT (name) DO NOTHING;