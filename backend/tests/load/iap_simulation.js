import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { ROOT_URL, API_URL, generateUserPayload, getAuthHeaders, generateStripeWebhook } from './config.js';

/**
 * iap_scenario.js
 * 
 * Beautifully simulate the full lifecycle of an IAP user:
 * 1. User registers
 * 2. User explores the app (access check - false)
 * 3. User subscribes (webhook simulation)
 * 4. User enjoys the app (access check - true)
 * 5. Admin monitors dashboard
 */

export const options = {
    scenarios: {
        user_lifecycle: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 50 }, // Ramp up to 50 active users
                { duration: '2m', target: 50 },  // Steady load
                { duration: '30s', target: 0 },  // Cool down
            ],
            gracefulRampDown: '30s',
        },
        admin_activity: {
            executor: 'constant-vus',
            vus: 5, // 5 active admins checking dashboards
            duration: '3m',
            startTime: '10s', // start slightly later
        },
        webhook_burst: {
            executor: 'constant-arrival-rate',
            rate: 2, // 2 purchase events per second
            timeUnit: '1s',
            duration: '3m',
            preAllocatedVUs: 50,
            maxVUs: 100,
        }

    },
    thresholds: {
        http_req_duration: ['p(95)<300'], // response time under 300ms
        http_req_failed: ['rate<0.01'],   // less than 1% errors
    },
};

export default function () {
    // Determine which scenario we are running (not strictly needed here but good for logic)
    // For simplicity, we just run a balanced default flow.

    group('1. Registration Flow', function () {
        const payload = generateUserPayload();
        const res = http.post(`${API_URL}/auth/register`, payload, {
            headers: { 'Content-Type': 'application/json' },
        });

        const success = check(res, {
            'registration status is 201': (r) => r.status === 201,
            'received access token': (r) => r.json('data.access_token') !== undefined,
        });

        if (!success) {
            console.error(`Registration failed for ${payload}`);
            return;
        }

        const token = res.json('data.access_token');
        const platform_id = JSON.parse(payload).platform_user_id;

        // 2. Initial Usage (Locked)
        group('2. Pre-subscription Usage', function () {
            const authRes = http.get(`${API_URL}/subscription/access`, getAuthHeaders(token));
            check(authRes, {
                'initial access status is 200': (r) => r.status === 200,
                'access is false initially': (r) => r.json('data.has_access') === false,
            });
            sleep(Math.random() * 2 + 1); // User stays 1-3 seconds
        });

        // 3. Purchase Event (Simulate Stripe Webhook)
        group('3. Simulation Purchase', function () {
            const webhookRes = http.post(`${ROOT_URL}/webhook/stripe`, generateStripeWebhook(platform_id), {
                headers: { 'Content-Type': 'application/json' },
            });
            check(webhookRes, {
                'webhook processed successfully': (r) => r.status === 200 || r.status === 204,
            });
            sleep(2); // Wait for worker to catch up (async processed)
        });

        // 4. Post-Purchase Usage (Unlocked)
        group('4. Premium Access Flow', function () {
            sleep(3); // Wait for worker to catch up (async processed)
            const premiumRes = http.get(`${API_URL}/subscription/access`, getAuthHeaders(token));
            check(premiumRes, {
                'access check is 200': (r) => r.status === 200,
                'access is now true': (r) => r.json('data.has_access') === true,
            });
        });

    });

    sleep(1);
}

// Separate function for admin simulation (if using scenarios with exec)
export function admin_behavior() {
    group('Admin Dashboard Usage', function () {
        // We'd normally use an admin token here, for now it hits public-ish ones
        const res = http.get(`${API_URL}/admin/dashboard/metrics`);
        check(res, {
            'admin dashboard status 200': (r) => r.status === 200 || r.status === 401, // 401 if unauthorized
        });
        sleep(5); // Admins check periodically
    });
}
