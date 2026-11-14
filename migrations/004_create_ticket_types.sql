CREATE TABLE IF NOT EXISTS ticket_types (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    price NUMERIC(12, 2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    quota INT NOT NULL DEFAULT 0 CHECK (quota >= 0),
    sold INT NOT NULL DEFAULT 0 CHECK (sold >= 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    CONSTRAINT uq_event_ticket_name UNIQUE (event_id, name),
    CONSTRAINT chk_sold_quota CHECK (sold <= quota)
);

CREATE INDEX IF NOT EXISTS idx_ticket_types_event ON ticket_types(event_id);
