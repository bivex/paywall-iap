import {Platform} from 'react-native';

// API Client configuration
const DEV_HOST = Platform.OS === 'android' ? 'http://10.0.2.2:8081' : 'http://localhost:8081';
export const API_BASE_URL = process.env.API_BASE_URL || (__DEV__ ? `${DEV_HOST}/v1` : 'https://api.yourapp.com/v1');
export const APP_ID = '2e0a62f9-dc32-4bcf-a2ce-8e545d9bbdf2';

export interface ApiResponse<T> {
  data: T;
  meta: {
    requestId: string;
    timestamp: string;
  };
}

export interface ApiError {
  error: string;
  message: string;
  code?: string;
  meta: {
    requestId: string;
    timestamp: string;
  };
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface RegisterRequest {
  platform_user_id: string;
  device_id: string;
  platform: 'ios' | 'android';
  app_version: string;
  email?: string;
}

export interface VerifyIAPRequest {
  platform: 'ios' | 'android';
  receipt_data: string;
  product_id: string;
  transaction_id?: string;
}
