-- 023_queue_payment_gate.sql
-- Gerbang pembayaran: tambah status MENUNGGU_BAYAR untuk antrian
-- yang sudah selesai diperiksa tapi belum melunasi tagihan.

ALTER TABLE queues DROP CONSTRAINT IF EXISTS queues_status_check;
ALTER TABLE queues ADD CONSTRAINT queues_status_check
    CHECK (status IN ('MENUNGGU', 'DIPANGGIL', 'SEDANG_DIPERIKSA', 'MENUNGGU_BAYAR', 'SELESAI', 'DIBATALKAN'));