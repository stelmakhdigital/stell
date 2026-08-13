import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 5,
  duration: "30s",
  thresholds: {
    http_req_duration: ["p(95)<5000"],
    http_req_failed: ["rate<0.05"],
  },
};

const BASE = __ENV.STELL_API || "http://127.0.0.1:8080";
const TOKEN = __ENV.STELL_API_TOKEN || "dev-token";

export default function () {
  const res = http.post(
    `${BASE}/v1/sessions`,
    JSON.stringify({ message: "ping" }),
    { headers: { Authorization: `Bearer ${TOKEN}`, "Content-Type": "application/json" } }
  );
  check(res, { "accepted or running": (r) => r.status === 202 || r.status === 200 });
  sleep(1);
}
