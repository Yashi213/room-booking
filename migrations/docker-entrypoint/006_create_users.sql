CREATE TABLE users (
    id             UUID PRIMARY KEY,
    email          VARCHAR(320) NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    role_id        SMALLINT NOT NULL REFERENCES roles (id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
