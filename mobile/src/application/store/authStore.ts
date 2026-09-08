import {create} from 'zustand';
import {persist, createJSONStorage} from 'zustand/middleware';
import {User} from '../../domain/entities/User';
import {AuthService} from '../../infrastructure/api/AuthService';
import {ApiClient} from '../../infrastructure/api/ApiClient';
import {SecureStorage} from '../../infrastructure/storage/SecureStorage';
import {getAuthService} from '../services/Services';

const getService = (): AuthService => {
  try {
    return getAuthService();
  } catch {
    return new AuthService(new ApiClient());
  }
};

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (deviceId: string, platform: 'ios' | 'android', appVersion: string, email?: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshAccessToken: () => Promise<string>;
  loadStoredTokens: () => Promise<void>;
  setUser: (user: User | null) => void;
  clearError: () => void;
}

export const isTokenExpired = (token: string): boolean => {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return true;
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const jsonStr = typeof atob === 'function' ? atob(base64) : Buffer.from(base64, 'base64').toString('utf8');
    const payload = JSON.parse(jsonStr);
    if (!payload.exp) return false;
    // 30 second safety margin
    return Date.now() >= payload.exp * 1000 - 30000;
  } catch {
    return true;
  }
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (email: string, password: string) => {
        set({isLoading: true, error: null});
        try {
          // TODO: Implement login with backend when endpoint is ready
          // For now, registration serves as login
          set({isLoading: false});
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Login failed',
          });
          throw error;
        }
      },

      register: async (deviceId: string, platform: 'ios' | 'android', appVersion: string, email?: string) => {
        set({isLoading: true, error: null});
        try {
          const authService = getService();
          const response = await authService.register(platform, deviceId, appVersion, email);

          try {
            const {getApiClient} = await import('../services/Services');
            getApiClient().setAccessToken(response.access_token);
          } catch {}

          const user: User = {
            id: response.user_id,
            platformUserId: deviceId,
            deviceId,
            platform,
            email,
            createdAt: new Date().toISOString(),
          };

          await SecureStorage.removeItem('user_logged_out');

          set({
            user,
            accessToken: response.access_token,
            refreshToken: response.refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Registration failed',
          });
          throw error;
        }
      },

      logout: async () => {
        set({isLoading: true, error: null});
        try {
          const refreshToken = get().refreshToken;
          const authService = getService();
          await authService.logout(refreshToken || undefined);

          // Mark user as explicitly logged out
          await SecureStorage.setItem('user_logged_out', 'true');

          // Reset subscriptions
          try {
            const {useSubscriptionStore} = await import('./subscriptionStore');
            useSubscriptionStore.getState().setSubscription(null);
          } catch {}

          // Reset IAP store
          try {
            const {useIAPStore} = await import('./iapStore');
            useIAPStore.getState().setCurrentPurchase(null);
          } catch {}

          // Reset Paywall SDK
          try {
            const {PaywallSDK} = await import('../../sdk');
            PaywallSDK.reset();
          } catch {}

          try {
            const {getApiClient} = await import('../services/Services');
            getApiClient().clearAccessToken();
          } catch {}

          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
          });
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Logout failed',
          });
        }
      },

      refreshAccessToken: async (): Promise<string> => {
        let token = get().refreshToken;
        if (!token) {
          token = (await SecureStorage.getItem('refresh_token')) || null;
        }
        if (!token) {
          throw new Error('No refresh token available');
        }

        set({isLoading: true, error: null});
        try {
          const authService = getService();
          const newTokens = await authService.refreshToken(token);

          try {
            const {getApiClient} = await import('../services/Services');
            getApiClient().setAccessToken(newTokens.access_token);
          } catch {}

          set({
            accessToken: newTokens.access_token,
            refreshToken: newTokens.refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });

          return newTokens.access_token;
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Token refresh failed',
          });
          throw error;
        }
      },

      loadStoredTokens: async () => {
        set({isLoading: true, error: null});
        try {
          const isExplicitlyLoggedOut = (await SecureStorage.getItem('user_logged_out')) === 'true';
          const authService = getService();
          const {accessToken, refreshToken} = await authService.getStoredTokens();

          if (accessToken && refreshToken && !isExplicitlyLoggedOut) {
            let validAccessToken = accessToken;
            if (isTokenExpired(accessToken)) {
              try {
                const refreshed = await authService.refreshToken(refreshToken);
                validAccessToken = refreshed.access_token;
                set({
                  accessToken: refreshed.access_token,
                  refreshToken: refreshed.refresh_token,
                  isAuthenticated: true,
                  isLoading: false,
                });
              } catch (refreshErr) {
                console.warn('[authStore] Refresh on load failed, attempting auto-registration:', refreshErr);
                const {default: DeviceInfo} = await import('react-native-device-info');
                const {Platform} = await import('react-native');
                const deviceId = await DeviceInfo.getUniqueId();
                const platform = Platform.OS === 'ios' ? 'ios' : 'android';
                const appVersion = DeviceInfo.getVersion();
                await get().register(deviceId, platform, appVersion);
                return;
              }
            } else {
              set({
                accessToken,
                refreshToken,
                isAuthenticated: true,
                isLoading: false,
              });
            }

            try {
              const {getApiClient} = await import('../services/Services');
              getApiClient().setAccessToken(validAccessToken);
            } catch {}
          } else if (!isExplicitlyLoggedOut) {
            // Auto-register device anonymously on first launch
            try {
              const {default: DeviceInfo} = await import('react-native-device-info');
              const {Platform} = await import('react-native');
              const deviceId = await DeviceInfo.getUniqueId();
              const platform = Platform.OS === 'ios' ? 'ios' : 'android';
              const appVersion = DeviceInfo.getVersion();
              await get().register(deviceId, platform, appVersion);
            } catch (regErr) {
              console.warn('[authStore] Auto-registration failed:', regErr);
              set({isLoading: false});
            }
          } else {
            set({
              user: null,
              accessToken: null,
              refreshToken: null,
              isAuthenticated: false,
              isLoading: false,
            });
          }
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Failed to load tokens',
          });
        }
      },

      setUser: (user: User | null) => set({user}),

      clearError: () => set({error: null}),
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => ({
        getItem: async (name: string) => {
          try {
            return await SecureStorage.getItem(name);
          } catch {
            return null;
          }
        },
        setItem: async (name: string, value: string) => {
          try {
            await SecureStorage.setItem(name, value);
          } catch (e) {
            console.error('Failed to save to secure storage:', e);
          }
        },
        removeItem: async (name: string) => {
          try {
            await SecureStorage.removeItem(name);
          } catch (e) {
            console.error('Failed to remove from secure storage:', e);
          }
        },
      })),
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        user: state.user,
      }),
    },
  ),
);
