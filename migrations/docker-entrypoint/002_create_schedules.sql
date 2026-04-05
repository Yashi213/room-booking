CREATE TABLE schedules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id       UUID NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    days_of_week  INTEGER[] NOT NULL,
    start_time    TEXT NOT NULL,
    end_time      TEXT NOT NULL,
    CONSTRAINT schedules_room_id_unique UNIQUE (room_id)
);
