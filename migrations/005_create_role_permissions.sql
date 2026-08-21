-- 005_create_role_permissions.sql
-- Tabel many-to-many antara roles dan permissions

CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission_id)
);

-- Index untuk pencarian berdasarkan role_id dan permission_id
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Insert default permissions untuk setiap role
-- ADMIN: semua permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'ADMIN'
ON CONFLICT DO NOTHING;

-- DOCTOR permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'DOCTOR' AND p.name IN (
    'patients.read',
    'queues.read',
    'queues.write',
    'medical_records.read',
    'medical_records.write',
    'prescriptions.read',
    'prescriptions.write'
)
ON CONFLICT DO NOTHING;

-- NURSE permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'NURSE' AND p.name IN (
    'patients.read',
    'patients.write',
    'registrations.read',
    'registrations.write',
    'queues.read',
    'queues.write',
    'medical_records.read',
    'medical_records.write'
)
ON CONFLICT DO NOTHING;

-- PHARMACIST permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'PHARMACIST' AND p.name IN (
    'prescriptions.read',
    'prescriptions.write',
    'medicines.read',
    'medicines.write',
    'pharmacy.read',
    'pharmacy.write'
)
ON CONFLICT DO NOTHING;

-- CASHIER permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'CASHIER' AND p.name IN (
    'payments.read',
    'payments.write',
    'patients.read'
)
ON CONFLICT DO NOTHING;

-- OWNER permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'OWNER' AND p.name IN (
    'reports.read',
    'settings.read'
)
ON CONFLICT DO NOTHING;
