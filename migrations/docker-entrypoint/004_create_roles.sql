CREATE TABLE roles (
    id   SMALLINT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
);

INSERT INTO roles (id, code) VALUES (1, 'admin'), (2, 'user');
