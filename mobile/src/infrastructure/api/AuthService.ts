import {ApiClient} from './ApiClient';
import {ApiResponse, AuthTokens, APP_ID} from './config';
import {SecureStorage} from '../storage/SecureStorage';

export interface RegisterRequest {
  platform_user_id: string;
  device_id: string;
  platform: 'ios' | 'android';
  app_version: string;
  email?: string;
  app_id?: string;
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
      app_id: APP_ID,
    };
    if (email && email.trim()) {
      request.email = email.trim();
    }

    try {
      const response = await this.api.post<ApiResponse<RegisterResponse>>('/auth/register', request);

      // Store tokens
      await this.storeTokens(response.data.access_token, response.data.refresh_token);
      this.api.setAccessToken(response.data.access_token);

      return response.data;
    } catch (err: any) {
      if (err?.error === 'CONFLICT' || err?.message?.includes('already exists')) {
        const stored = await this.getStoredTokens();
        if (stored.accessToken && stored.refreshToken) {
          this.api.setAccessToken(stored.accessToken);
          return {
            user_id: deviceId,
            access_token: stored.accessToken,
            refresh_token: stored.refreshToken,
            expires_in: 900,
          };
        }
      }
      throw err;
    }
  }

  async refreshToken(refreshToken: string): Promise<void> {
    const response = await this.api.post<ApiResponse<{access_token: string; refresh_token: string; expires_in: number}>>('/auth/refresh', {
      refresh_token: refreshToken,
    });

    await this.storeTokens(response.data.access_token, response.data.refresh_token);
    this.api.setAccessToken(response.data.access_token);
  }

  async logout(refreshToken?: string): Promise<void> {
    try {
      if (refreshToken) {
        await this.api.post<ApiResponse<void>>('/auth/logout', {
          refresh_token: refreshToken,
        });
      }
    } catch {
      // Best-effort remote revocation
    } finally {
      await SecureStorage.removeItem('access_token');
      await SecureStorage.removeItem('refresh_token');
      this.api.clearAccessToken();
    }
  }

  private async storeTokens(accessToken: string, refreshToken: string): Promise<void> {
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
