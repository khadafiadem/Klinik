CREATE TABLE IF NOT EXISTS pain_assessments (
    id SERIAL PRIMARY KEY,
    medical_record_id INTEGER NOT NULL UNIQUE REFERENCES medical_records(id),
    location TEXT,
    quality VARCHAR(20),
    intensity INTEGER NOT NULL DEFAULT 0 CHECK (intensity BETWEEN 0 AND 10),
    onset_duration TEXT,
    radiation TEXT,
    aggravating_factors TEXT,
    relieving_factors TEXT,
    notes TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pain_assessments_mr ON pain_assessments(medical_record_id);