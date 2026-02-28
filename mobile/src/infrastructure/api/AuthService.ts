import {ApiClient, ApiResponse, AuthTokens} from './ApiClient';

export interface RegisterRequest {
  platform_user_id: string;
  device_id: string;
  platform: 'ios' | 'android';
  app_version: string;
  email?: string;
}

export interface RegisterResponse {
  user_id: string;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export class AuthService {
  constructor(private api: ApiClient) {}

  async register(platform: 'ios' | 'android', deviceId: string, appVersion: string, email?: string): Promise<RegisterResponse> {
    const request: RegisterRequest = {
      platform_user_id: this.getPlatformUserId(platform, deviceId),
      device_id: deviceId,
      platform,
      app_version: appVersion,
      email,
    };

    const response = await this.api.post<ApiResponse<RegisterResponse>>('/auth/register', request);

    // Store tokens
    await this.storeTokens(response.data.access_token, response.data.refresh_token);
    this.api.setAccessToken(response.data.access_token);

    return response.data;
  }

  async refreshToken(refreshToken: string): Promise<void> {
    const response = await this.api.post<ApiResponse<{access_token: string; refresh_token: string; expires_in: number}>>('/auth/refresh', {
      refresh_token: refreshToken,
    });

    await this.storeTokens(response.data.access_token, response.data.refresh_token);
    this.api.setAccessToken(response.data.access_token);
  }

  private async storeTokens(accessToken: string, refreshToken: string): Promise<void> {
    const SecureStorage = require('react-native-secure-storage').default;
    await SecureStorage.setItem('access_token', accessToken);
    await SecureStorage.setItem('refresh_token', refreshToken);
  }

  private getPlatformUserId(platform: string, deviceId: string): string {
    if (platform === 'ios') {
      // iOS: Vendor ID (unique per device, persists until app uninstall)
      return deviceId;
    } else {
      // Android: Google obfuscated account ID
      // This is just a placeholder - real implementation uses Google Play Services
      return `android_${deviceId}`;
    }
  }

  async getStoredTokens(): Promise<{accessToken: string | null; refreshToken: string | null}> {
    const SecureStorage = require('react-native-secure-storage').default;
    try {
      const accessToken = await SecureStorage.getItem('access_token');
      const refreshToken = await SecureStorage.getItem('refresh_token');
      return {accessToken, refreshToken};
    } catch (error) {
      console.error('Failed to get tokens:', error);
      return {accessToken: null, refreshToken: null};
    }
  }
}
