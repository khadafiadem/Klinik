CREATE TABLE IF NOT EXISTS immunization_schedules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    age_label VARCHAR(50) NOT NULL,
    month_from INT NOT NULL DEFAULT 0,
    month_to INT,
    category VARCHAR(20) NOT NULL DEFAULT 'WAJIB',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS patient_immunizations (
    id SERIAL PRIMARY KEY,
    patient_id INTEGER NOT NULL REFERENCES patients(id),
    immunization_id INTEGER REFERENCES immunization_schedules(id),
    vaccination_date DATE NOT NULL,
    vaccine_name VARCHAR(100) NOT NULL,
    batch_number VARCHAR(50),
    provider_name VARCHAR(100),
    notes TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patient_imm_patient ON patient_immunizations(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_imm_date ON patient_immunizations(vaccination_date);

-- Jadwal Imunisasi IDAI (acuan umum)
INSERT INTO immunization_schedules (name, age_label, month_from, month_to, category, sort_order) VALUES
    ('Hepatitis B 1',                'Baru lahir (0-24 jam)',    0, 0,  'WAJIB',  1),
    ('BCG',                          '1 bulan',                  1, 1,  'WAJIB',  2),
    ('Polio 1 (OPV)',                '1 bulan',                  1, 1,  'WAJIB',  3),
    ('DTP-HB-Hib 1',                 '2 bulan',                  2, 2,  'WAJIB',  4),
    ('Polio 2 (OPV)',                '2 bulan',                  2, 2,  'WAJIB',  5),
    ('PCV 1',                        '2 bulan',                  2, 2,  'WAJIB',  6),
    ('Rotavirus 1',                  '2 bulan',                  2, 2,  'WAJIB',  7),
    ('DTP-HB-Hib 2',                 '3 bulan',                  3, 3,  'WAJIB',  8),
    ('Polio 3 (OPV)',                '3 bulan',                  3, 3,  'WAJIB',  9),
    ('Rotavirus 2',                  '3 bulan',                  3, 3,  'WAJIB',  10),
    ('DTP-HB-Hib 3',                 '4 bulan',                  4, 4,  'WAJIB',  11),
    ('Polio 4 (IPV)',                '4 bulan',                  4, 4,  'WAJIB',  12),
    ('PCV 2',                        '4 bulan',                  4, 4,  'WAJIB',  13),
    ('PCV 3',                        '12 bulan',                 12, 12, 'WAJIB',  14),
    ('Influenza 1',                  '6 bulan',                  6, 6,  'ANJURAN', 15),
    ('MR 1',                         '9 bulan',                  9, 9,  'WAJIB',  16),
    ('Influenza 2',                  '7 bulan',                  7, 7,  'ANJURAN', 17),
    ('MMR',                          '15 bulan',                 15, 15, 'WAJIB',  18),
    ('DTP-HB-Hib 4',                 '18 bulan',                 18, 18, 'WAJIB',  19),
    ('Varicella 1',                  '12 bulan',                 12, 12, 'ANJURAN', 20),
    ('JE (daerah endemis)',          '9 bulan',                  9, 9,  'ANJURAN', 21),
    ('Typhoid',                      '2 tahun (ulang tiap 3 th)', 24, 24, 'ANJURAN', 22),
    ('Hepatitis A',                  '2 tahun',                  24, 24, 'ANJURAN', 23),
    ('MMR 2',                        '5-7 tahun',                60, 84,  'WAJIB',  24),
    ('DTP (booster) + Polio',        '5-7 tahun',                60, 84,  'WAJIB',  25),
    ('HPV',                          '9-14 tahun',               108, 168, 'ANJURAN', 26)
ON CONFLICT (name) DO NOTHING;