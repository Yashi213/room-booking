from __future__ import annotations

import os
import sys
import uuid
from contextlib import contextmanager
from datetime import datetime, timedelta, timezone
from itertools import cycle
from urllib.parse import parse_qs, unquote, urlparse

os.environ.setdefault("PGCLIENTENCODING", "UTF8")

import psycopg2
from psycopg2.extras import execute_values


DEFAULT_DATABASE_URL = (
    "postgres://postgres:postgres@127.0.0.1:5432/roombooking?sslmode=disable"
)
DEFAULT_DATABASE_URL_DOCKER = (
    "postgres://postgres:postgres@postgres:5432/roombooking?sslmode=disable"
)
K6_ROOM_ID = uuid.UUID("a0000001-0000-4000-8000-000000000001")
DUMMY_PASSWORD_HASH = "$2a$10$seedNotForRealLoginHashPlaceholder"


def _postgres_kwargs(dsn: str) -> dict:
    u = urlparse(dsn)

    dbname = (u.path or "").lstrip("/")

    kw: dict = {
        "host": u.hostname or "localhost",
        "port": u.port or 5432,
        "user": unquote(u.username or ""),
        "password": unquote(u.password or ""),
        "dbname": unquote(dbname),
        "connect_timeout": 10,
        "options": "-c client_encoding=UTF8",
    }
    for k, vals in parse_qs(u.query).items():
        if k == "sslmode" and vals:
            kw["sslmode"] = vals[0]
    return kw


@contextmanager
def _without_database_url():
    s = os.environ.pop("DATABASE_URL", None)
    try:
        yield
    finally:
        if s is not None:
            os.environ["DATABASE_URL"] = s


def _default_database_url() -> str:
    if os.path.exists("/.dockerenv"):
        return DEFAULT_DATABASE_URL_DOCKER
    return DEFAULT_DATABASE_URL


def connect():
    use_libpq = os.environ.get("SEED_USE_LIBPQ_ENV", "").lower() in ("1", "true", "yes")
    dsn = os.environ.get("DATABASE_URL") or _default_database_url()
    with _without_database_url():
        if use_libpq:
            host = os.environ.get("PGHOST")
            if not host and os.path.exists("/.dockerenv"):
                host = "postgres"
            return psycopg2.connect(
                host=host or "127.0.0.1",
                port=int(os.environ.get("PGPORT", "5432")),
                user=os.environ.get("PGUSER", "postgres"),
                password=os.environ.get("PGPASSWORD", "postgres"),
                dbname=os.environ.get("PGDATABASE", "roombooking"),
                sslmode=os.environ.get("PGSSLMODE", "disable"),
                connect_timeout=10,
                options="-c client_encoding=UTF8",
            )
        return psycopg2.connect(**_postgres_kwargs(dsn))



def tomorrow_midnight() -> datetime:
    t = datetime.now(timezone.utc).date() + timedelta(days=1)
    return datetime(t.year, t.month, t.day, 0, 0, 0, tzinfo=timezone.utc)


def max_slots_from_start(start_hm: str) -> int:
    sh, sm = map(int, start_hm.split(":"))
    start_m = sh * 60 + sm
    return (24 * 60 - start_m) // 30


def per_room_counts(total: int, rooms: int) -> list[int]:
    b, e = total // rooms, total % rooms
    return [b + (1 if i < e else 0) for i in range(rooms)]


def build_all_slots(
    room_ids: list[uuid.UUID],
    day0: datetime,
    start_hm: str,
    counts: list[int],
) -> list[tuple[uuid.UUID, uuid.UUID, datetime, datetime]]:
    sh, sm = map(int, start_hm.split(":"))
    base = day0.replace(hour=sh, minute=sm, second=0, microsecond=0)
    out = []
    for rid, n in zip(room_ids, counts):
        for i in range(n):
            a = base + timedelta(minutes=30 * i)
            b = a + timedelta(minutes=30)
            out.append((uuid.uuid4(), rid, a, b))
    return out


def schedule_end_hm(start_hm: str, n_slots: int) -> str:
    sh, sm = map(int, start_hm.split(":"))
    end_m = sh * 60 + sm + 30 * n_slots
    if end_m > 24 * 60:
        return "23:59"
    eh, em = end_m // 60, end_m % 60
    return f"{eh:02d}:{em:02d}"


def main() -> int:
    rooms_n = int(os.environ.get("SEED_ROOMS", "50"))
    slots_day = int(os.environ.get("SEED_SLOTS_PER_DAY", "1000"))
    users_n = int(os.environ.get("SEED_USERS", "10000"))
    bookings_n = int(os.environ.get("SEED_BOOKINGS", "100000"))
    start_hm = os.environ.get("SEED_SLOT_START", "08:00")



    counts = per_room_counts(slots_day, rooms_n)
    need = max(counts)
    if need > max_slots_from_start(start_hm):
        return 1

    end_hm = schedule_end_hm(start_hm, need)
    day0 = tomorrow_midnight()

    room_ids = [K6_ROOM_ID] + [uuid.uuid4() for _ in range(rooms_n - 1)]

    conn = connect()
    try:
        cur = conn.cursor()
        cur.execute("DELETE FROM bookings")
        cur.execute("DELETE FROM slots")
        cur.execute("DELETE FROM schedules")
        cur.execute("DELETE FROM rooms")
        cur.execute("DELETE FROM users")

        execute_values(
            cur,
            "INSERT INTO rooms (id, name, description, capacity) VALUES %s",
            [
                (str(rid), f"room-{i}", f"seed room {i}", 10 + (i % 20))
                for i, rid in enumerate(room_ids)
            ],
            template="(%s::uuid, %s, %s, %s)",
        )

        days = list(range(1, 8))
        execute_values(
            cur,
            """
            INSERT INTO schedules (room_id, days_of_week, start_time, end_time)
            VALUES %s
            """,
            [
                (str(rid), days, start_hm, end_hm) for rid in room_ids
            ],
            template="(%s::uuid, %s::int[], %s, %s)",
        )

        slot_rows = build_all_slots(room_ids, day0, start_hm, counts)
        execute_values(
            cur,
            "INSERT INTO slots (id, room_id, start_time, end_time) VALUES %s",
            [(str(sid), str(rid), a, b) for sid, rid, a, b in slot_rows],
            template="(%s::uuid, %s::uuid, %s::timestamptz, %s::timestamptz)",
        )

        slot_ids = [r[0] for r in slot_rows]
        user_rows = []
        user_ids = []
        for i in range(users_n):
            uid = uuid.uuid4()
            user_ids.append(uid)
            user_rows.append(
                (str(uid), f"seed_{i}@seed.local", DUMMY_PASSWORD_HASH, 2)
            )
        execute_values(
            cur,
            "INSERT INTO users (id, email, password_hash, role_id) VALUES %s",
            user_rows,
            template="(%s::uuid, %s, %s, %s)",
        )

        n_slots = len(slot_ids)
        n_active = min(n_slots, bookings_n)
        n_cancel = bookings_n - n_active

        uc = cycle(user_ids)
        booking_rows = []
        for sid in slot_ids[:n_active]:
            booking_rows.append(
                (str(uuid.uuid4()), str(sid), str(next(uc)), 1, None)
            )
        sid_c = cycle(slot_ids)
        for _ in range(n_cancel):
            booking_rows.append(
                (str(uuid.uuid4()), str(next(sid_c)), str(next(uc)), 2, None)
            )

        batch = 5000
        for i in range(0, len(booking_rows), batch):
            chunk = booking_rows[i : i + batch]
            execute_values(
                cur,
                """
                INSERT INTO bookings (id, slot_id, user_id, status_id, conference_link)
                VALUES %s
                """,
                chunk,
                template="(%s::uuid, %s::uuid, %s::uuid, %s, %s)",
            )

        conn.commit()
        cur.close()
    finally:
        conn.close()

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
