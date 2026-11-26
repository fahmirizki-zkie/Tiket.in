-- Migration 009: Manajemen user oleh admin (Sprint 2 - Fatin)
-- Menambah kolom updated_at (dipakai update profil, ganti password, ubah role)
-- dan index role untuk filter GET /users?role=...

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);