import http from "k6/http";
import { check, sleep } from "k6";

const BASE = __ENV.BASE_URL || "http://localhost:8080";
const ROOM_ID =
  __ENV.ROOM_ID || "a0000001-0000-4000-8000-000000000001";
const RPS = Number(__ENV.RPS || 100);
const DURATION = __ENV.DURATION || "60s";
const P95_MS = __ENV.P95_MS || "200";
const FAIL_MAX = __ENV.FAIL_RATE || "0.001";

function tomorrow() {
  const d = new Date();
  d.setUTCDate(d.getUTCDate() + 1);
  return d.toISOString().slice(0, 10);
}

export function setup() {
  const login = http.post(
    `${BASE}/dummyLogin`,
    JSON.stringify({ role: "user" }),
    { headers: { "Content-Type": "application/json" } },
  );
  if (login.status !== 200) {
    throw new Error(`dummyLogin failed: ${login.status} ${login.body}`);
  }
  const token = login.json("token");
  const date = tomorrow();
  const probe = http.get(
    `${BASE}/rooms/${ROOM_ID}/slots/list?date=${date}`,
    { headers: { Authorization: `Bearer ${token}` } },
  );
  if (probe.status !== 200) {
    const body =
      typeof probe.body === "string"
        ? probe.body.slice(0, 500)
        : String(probe.body);
    throw new Error(
      `GET slots/list failed (${probe.status}): ${body}. Сначала: python scripts/seed.py`,
    );
  }
  return { token };
}

export const options = {
  scenarios: {
    slots_list: {
      executor: "constant-arrival-rate",
      rate: RPS,
      timeUnit: "1s",
      duration: DURATION,
      preAllocatedVUs: Number(__ENV.PREALLOC_VUS || Math.min(200, RPS * 3)),
      maxVUs: Number(__ENV.MAX_VUS || 500),
    },
  },
  thresholds: {
    http_req_failed: [`rate<${FAIL_MAX}`],
    http_req_duration: [`p(95)<${P95_MS}`],
  },
};

export default function (data) {
  const date = tomorrow();
  const res = http.get(
    `${BASE}/rooms/${ROOM_ID}/slots/list?date=${date}`,
    { headers: { Authorization: `Bearer ${data.token}` } },
  );
  check(res, { "status 200": (r) => r.status === 200 });
  sleep(0.001);
}
