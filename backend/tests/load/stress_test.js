import http from 'k6/http';
import { check, sleep } from 'k6';
import { API_URL, generateUserPayload, getAuthHeaders } from './config.js';

export const options = {
    stages: [
        { duration: '30s', target: 50 },  // Ramp up to 50 virtual users
        { duration: '1m', target: 100 },  // Ramp up to 100 virtual users (high load)
        { duration: '2m', target: 200 },  // Push to limits
        { duration: '30s', target: 0 },   // Cool down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // Allowing slightly slower responses under stress
        http_req_failed: ['rate<0.05'],   // Max 5% failure rate under high stress
    },
};

export default function () {
    const regPayload = generateUserPayload();
    const regRes = http.post(`${API_URL}/auth/register`, regPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    if (regRes.status !== 201) {
        return; // Don't continue if registration failed
    }

    const token = regRes.json('data.access_token');
    const authHeaders = getAuthHeaders(token);

    sleep(0.5); // Throttling intentionally reduced to stress test API mapping 

    const accessRes = http.get(`${API_URL}/subscription/access`, authHeaders);

    check(accessRes, {
        'access check completed': (r) => r.status === 200,
    });
}
