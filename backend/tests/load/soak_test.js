import http from 'k6/http';
import { check, sleep } from 'k6';
import { API_URL, generateUserPayload, getAuthHeaders } from './config.js';

export const options = {
    stages: [
        { duration: '2m', target: 20 },  // Ramp up to normal load
        { duration: '1h', target: 20 },  // Sustain normal load for 1 hour
        { duration: '2m', target: 0 },   // Cool down
    ],
    thresholds: {
        http_req_duration: ['p(99)<300'], // 99% of requests < 300ms
        http_req_failed: ['rate<0.001'],  // Minimal failures allowed
    },
};

export default function () {
    const regPayload = generateUserPayload();
    const regRes = http.post(`${API_URL}/auth/register`, regPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    if (regRes.status !== 201) return;

    const token = regRes.json('data.access_token');
    const authHeaders = getAuthHeaders(token);

    sleep(1);

    const accessRes = http.get(`${API_URL}/subscription/access`, authHeaders);

    check(accessRes, {
        'access check completed': (r) => r.status === 200,
    });

    sleep(1);
}
