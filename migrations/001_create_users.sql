CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL    PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role        VARCHAR(20)  NOT NULL DEFAULT 'peserta'
                CHECK (role IN ('peserta', 'panitia', 'admin')),
    created_at  TIMESTAMP    NOT NULL DEFAULT now()
);
