CREATE TABLE IF NOT EXISTS letter_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    subject VARCHAR(200),
    body_template TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS letter_issuances (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES letter_templates(id),
    letter_number VARCHAR(50) NOT NULL,
    patient_id INTEGER NOT NULL REFERENCES patients(id),
    doctor_id INTEGER REFERENCES doctors(id),
    notes TEXT,
    issued_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_letter_issuances_letter ON letter_issuances(letter_number);

INSERT INTO letter_templates (name, code, subject, body_template) VALUES
    ('Surat Keterangan Sehat', 'SEHAT', 'SURAT KETERANGAN SEHAT',
     'Yang bertanda tangan di bawah ini, dokter {{dokter}} dari {{klinik}}, menerangkan bahwa:\n\nNama : {{nama}}\nNIK : {{nik}}\nTempat, Tgl. Lahir : {{ttl}}\nJenis Kelamin : {{jk}}\nAlamat : {{alamat}}\n\nBerdasarkan hasil pemeriksaan kesehatan yang dilakukan pada {{tanggal}}, yang bersangkutan dinyatakan SEHAT dan layak untuk melakukan aktivitas sehari-hari.'),
    ('Surat Keterangan Sakit', 'SAKIT', 'SURAT KETERANGAN SAKIT',
     'Yang bertanda tangan di bawah ini, dokter {{dokter}} dari {{klinik}}, menerangkan bahwa:\n\nNama : {{nama}}\nNIK : {{nik}}\nTempat, Tgl. Lahir : {{ttl}}\nJenis Kelamin : {{jk}}\nAlamat : {{alamat}}\n\nBahwa yang bersangkutan sedang dalam keadaan sakit dan membutuhkan istirahat. Surat ini diberikan untuk dipergunakan sebagaimana mestinya.'),
    ('Surat Rujukan', 'RUJUKAN', 'SURAT RUJUKAN',
     'Kepada Yth. :\nKepala FKTP / Fasilitas Kesehatan\n\nPerihal : Permohonan Pemeriksaan Lanjutan\n\nDengan hormat, bersama ini kami rujuk pasien kami:\n\nNama : {{nama}}\nNIK : {{nik}}\nTempat, Tgl. Lahir : {{ttl}}\nJenis Kelamin : {{jk}}\nNo. RM : {{no_rm}}\nAlamat : {{alamat}}\n\nuntuk mendapatkan pemeriksaan / penanganan lebih lanjut sesuai indikasi medis. Demikian surat rujukan ini kami buat, atas perhatian dan kerjasamanya kami ucapkan terima kasih.')
ON CONFLICT (code) DO NOTHING;