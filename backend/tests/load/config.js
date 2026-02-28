// Common setup for all k6 tests
export const ROOT_URL = __ENV.ROOT_URL || 'http://localhost:8080';
export const API_URL = `${ROOT_URL}/v1`;

export function getAuthHeaders(token) {
    return {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        }
    };
}

export function generateUserPayload() {
    const timestamp = Date.now();
    const randomStr = Math.random().toString(36).substring(7);
    return JSON.stringify({
        platform_user_id: `k6_user_${timestamp}_${randomStr}`,
        device_id: `k6_device_${randomStr}`,
        platform: 'ios',
        app_version: '1.0.0',
        email: `k6_${timestamp}_${randomStr}@example.com`
    });
}

export function generateStripeWebhook(platform_user_id) {
    return JSON.stringify({
        id: `evt_${Math.random().toString(36).substring(7)}`,
        type: 'invoice.payment_succeeded',
        data: {
            object: {
                customer: platform_user_id, // assuming we map this
                subscription: `sub_${Math.random().toString(36).substring(7)}`,
                amount_paid: 999,
                currency: 'usd'
            }
        }
    });
}

