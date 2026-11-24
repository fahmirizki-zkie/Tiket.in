-- Migration 007: Outbox Table & Partial Unique Index for Registration Cancellation
-- Memperbaiki constraint pendaftaran ganda agar tiket yang dibatalkan mengizinkan pendaftaran ulang
-- Menambahkan tabel outboxes untuk pola Transactional Outbox (Modul 11 & 12)

-- 1. Hapus constraint pendaftaran lama jika ada
ALTER TABLE registrations DROP CONSTRAINT IF EXISTS uq_registrations_user_event;

-- 2. Buat Partial Unique Index: Mencegah pendaftaran ganda HANYA untuk tiket yang masih aktif (bukan cancelled)
CREATE UNIQUE INDEX IF NOT EXISTS uq_registrations_active_user_event 
ON registrations (user_id, event_id) 
WHERE status != 'cancelled';

-- 3. Tabel Outbox untuk pola Transactional Outbox
CREATE TABLE IF NOT EXISTS outboxes (
    id          BIGSERIAL    PRIMARY KEY,
    event_type  VARCHAR(100) NOT NULL,
    payload     JSONB        NOT NULL,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    retry_count INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_outbox_status_created ON outboxes(status, created_at);
