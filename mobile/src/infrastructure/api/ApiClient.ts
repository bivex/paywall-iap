import {API_BASE_URL, APP_ID} from './config';

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

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({
        error: 'UNKNOWN_ERROR',
        message: `HTTP ${response.status}`,
      }));
      throw error;
    }

    return response.json();
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
