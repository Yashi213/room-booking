CREATE TABLE booking_statuses (
    id   SMALLINT PRIMARY KEY,
    code TEXT NOT NULL UNIQUE
);

INSERT INTO booking_statuses (id, code) VALUES (1, 'active'), (2, 'cancelled');
