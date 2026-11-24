-- Migration 008: Add cancellation metadata to events table
-- Praktikum Pemrograman Backend Lanjut - D4 Teknik Informatika UNAIR

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS cancellation_reason TEXT,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_events_cancelled_at ON events(cancelled_at) WHERE status = 'cancelled';
