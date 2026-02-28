import http from 'k6/http';
import { check, sleep } from 'k6';
import { API_URL, generateUserPayload, getAuthHeaders } from './config.js';

export const options = {
    stages: [
        { duration: '30s', target: 20 }, // Ramp up to 20 virtual users over 30 seconds
        { duration: '1m', target: 20 },  // Stay at 20 virtual users for 1 minute
        { duration: '30s', target: 0 },  // Ramp down to 0 virtual users over 30 seconds
    ],
    thresholds: {
        http_req_duration: ['p(95)<200'], // 95% of requests must complete within 200ms
        http_req_failed: ['rate<0.01'],   // Error rate must be < 1%
    },
};

export default function () {
    // 1. Register a new user
    const regPayload = generateUserPayload();
    const regRes = http.post(`${API_URL}/auth/register`, regPayload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(regRes, {
        'registration successful': (r) => r.status === 201,
        'has access token': (r) => r.json('data.access_token') !== undefined,
    });

    if (regRes.status !== 201) {
        return; // Don't continue if registration failed
    }

    const token = regRes.json('data.access_token');
    const authHeaders = getAuthHeaders(token);

    // Give the server a tiny breather
    sleep(1);

    // 2. Check access (should be false since no subscription yet)
    const accessRes = http.get(`${API_URL}/subscription/access`, authHeaders);

    check(accessRes, {
        'access check successful': (r) => r.status === 200,
        'access is false': (r) => r.json('data.has_access') === false,
    });

    sleep(1);
}
