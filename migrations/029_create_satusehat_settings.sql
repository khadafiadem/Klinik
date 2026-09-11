-- 029_create_satusehat_settings.sql
-- Pengaturan integrasi SATUSEHAT (single row)

CREATE TABLE IF NOT EXISTS satusehat_settings (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(10) NOT NULL DEFAULT 'PROD',
    base_url VARCHAR(255),
    org_id VARCHAR(50),
    org_name VARCHAR(200),
    location_id VARCHAR(50),
    client_id VARCHAR(200),
    client_secret TEXT,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO satusehat_settings (environment, base_url)
SELECT 'PROD', 'https://api-satusehat.kemkes.go.id'
WHERE NOT EXISTS (SELECT 1 FROM satusehat_settings);

CREATE OR REPLACE FUNCTION update_satusehat_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_satusehat_settings_updated_at
    BEFORE UPDATE ON satusehat_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_satusehat_settings_updated_at();