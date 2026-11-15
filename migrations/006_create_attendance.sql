CREATE TABLE IF NOT EXISTS attendances (
    id BIGSERIAL PRIMARY KEY,
    registration_id BIGINT NOT NULL UNIQUE REFERENCES registrations(id) ON DELETE RESTRICT,
    checked_in_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    checked_in_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_attendances_registration_id ON attendances(registration_id);
CREATE INDEX IF NOT EXISTS idx_attendances_checked_in_by ON attendances(checked_in_by);
