import {API_BASE_URL, APP_ID, ApiResponse} from './config';

export type {ApiResponse};

export class ApiClient {
  private accessToken: string | null = null;
  private appId: string = APP_ID;

  setAccessToken(token: string) {
    this.accessToken = token;
  }

  clearAccessToken() {
    this.accessToken = null;
  }

  setAppId(appId: string) {
    this.appId = appId;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-App-ID': this.appId,
    };

    let token = this.accessToken;
    if (!token) {
      try {
        const {SecureStorage} = await import('../storage/SecureStorage');
        token = await SecureStorage.getItem('access_token');
        if (token) {
          this.accessToken = token;
        }
      } catch {
        // ignore
      }
    }

    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    let response = await fetch(url, {
      ...options,
      headers,
    });

    if (response.status === 401 && !endpoint.startsWith('/auth/')) {
      try {
        const {useAuthStore} = await import('../../application/store/authStore');
        const newToken = await useAuthStore.getState().refreshAccessToken();
        if (newToken) {
          headers['Authorization'] = `Bearer ${newToken}`;
          response = await fetch(url, {
            ...options,
            headers,
          });
        }
      } catch {
        try {
          const {SecureStorage} = await import('../storage/SecureStorage');
          const isExplicitlyLoggedOut = (await SecureStorage.getItem('user_logged_out')) === 'true';
          if (!isExplicitlyLoggedOut) {
            const {useAuthStore} = await import('../../application/store/authStore');
            const {default: DeviceInfo} = await import('react-native-device-info');
            const {Platform} = await import('react-native');
            const deviceId = await DeviceInfo.getUniqueId();
            const platform = Platform.OS === 'ios' ? 'ios' : 'android';
            const appVersion = DeviceInfo.getVersion();
            await useAuthStore.getState().register(deviceId, platform, appVersion);
            const freshToken = useAuthStore.getState().accessToken;
            if (freshToken) {
              headers['Authorization'] = `Bearer ${freshToken}`;
              response = await fetch(url, {
                ...options,
                headers,
              });
            }
          }
        } catch {
          // ignore
        }
      }
    }

    if (!response.ok) {
      let errorData: any = null;
      try {
        const text = await response.text();
        if (text && text.trim()) {
          errorData = JSON.parse(text);
        }
      } catch {
        // ignore
      }
      const error = errorData || {
        error: 'UNKNOWN_ERROR',
        message: `HTTP ${response.status}`,
      };
      throw error;
    }

    if (response.status === 204 || response.headers.get('content-length') === '0') {
      return {} as T;
    }

    const text = await response.text();
    if (!text || !text.trim()) {
      return {} as T;
    }

    try {
      return JSON.parse(text);
    } catch {
      return {} as T;
    }
  }

  async post<T>(endpoint: string, data: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'GET',
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'DELETE',
    });
  }
}
