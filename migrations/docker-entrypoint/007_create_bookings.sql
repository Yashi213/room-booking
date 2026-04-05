CREATE TABLE bookings (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_id          UUID NOT NULL REFERENCES slots (id) ON DELETE CASCADE,
    user_id          UUID NOT NULL,
    status_id        SMALLINT NOT NULL REFERENCES booking_statuses (id),
    conference_link  TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX bookings_one_active_per_slot
    ON bookings (slot_id)
    WHERE status_id = 1;

CREATE INDEX bookings_user_id_idx ON bookings (user_id);
