import { SharedArray } from 'k6/data';

// Common setup for all k6 tests
export const API_URL = __ENV.BASE_URL || 'http://localhost:8080/v1';

export function getAuthHeaders(token) {
    return {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
        }
    };
}

export function generateUserPayload() {
    const timestamp = new Date().getTime();
    const randomStr = Math.random().toString(36).substring(7);
    return JSON.stringify({
        platform_user_id: `k6_user_${timestamp}_${randomStr}`,
        device_id: `k6_device_${randomStr}`,
        platform: 'ios',
        app_version: '1.0.0',
        email: `k6_${timestamp}_${randomStr}@example.com`
    });
}
