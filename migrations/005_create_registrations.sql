CREATE TABLE IF NOT EXISTS registrations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    event_id BIGINT NOT NULL REFERENCES events(id),
    ticket_type_id BIGINT NOT NULL REFERENCES ticket_types(id),
    ticket_code VARCHAR(20) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'registered',
    registered_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT chk_registrations_status
        CHECK (status IN ('registered', 'cancelled', 'checked_in')),
    CONSTRAINT uq_registrations_user_event
        UNIQUE (user_id, event_id)
);

CREATE INDEX IF NOT EXISTS idx_registrations_user_id ON registrations(user_id);
CREATE INDEX IF NOT EXISTS idx_registrations_ticket_type_id ON registrations(ticket_type_id);