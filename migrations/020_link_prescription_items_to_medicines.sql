-- 020_link_prescription_items_to_medicines.sql
-- Hubungkan item resep dengan master obat agar stok berkurang saat resep diselesaikan apotek.

ALTER TABLE prescription_items ADD COLUMN IF NOT EXISTS medicine_id INTEGER REFERENCES medicines(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_prescription_items_medicine ON prescription_items(medicine_id);
