// API Client configuration
const API_BASE_URL = __DEV__ ? 'http://localhost:8080/v1' : 'https://api.yourapp.com/v1';

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
