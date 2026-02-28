import {create} from 'zustand';
import {persist, createJSONStorage} from 'zustand/middleware';
import {User} from '../../domain/entities/User';
import {AuthService} from '../../infrastructure/api/AuthService';
import {SecureStorage} from '../../infrastructure/storage/SecureStorage';

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
  refreshAccessToken: () => Promise<void>;
  loadStoredTokens: () => Promise<void>;
  setUser: (user: User | null) => void;
  clearError: () => void;
}

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
          const authService = new AuthService(/* apiClient */);
          const response = await authService.register(platform, deviceId, appVersion, email);

          const user: User = {
            id: response.user_id,
            platformUserId: deviceId,
            deviceId,
            platform,
            email,
            createdAt: new Date().toISOString(),
          };

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
          await SecureStorage.removeItem('access_token');
          await SecureStorage.removeItem('refresh_token');

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

      refreshAccessToken: async () => {
        const {refreshToken: token} = get();
        if (!token) {
          throw new Error('No refresh token available');
        }

        set({isLoading: true, error: null});
        try {
          const authService = new AuthService(/* apiClient */);
          await authService.refreshToken(token);

          set({isLoading: false});
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
          const authService = new AuthService(/* apiClient */);
          const {accessToken, refreshToken} = await authService.getStoredTokens();

          if (accessToken && refreshToken) {
            set({
              accessToken,
              refreshToken,
              isAuthenticated: true,
              isLoading: false,
            });
          } else {
            set({isLoading: false});
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
            const SecureStorage = require('react-native-secure-storage').default;
            return await SecureStorage.getItem(name);
          } catch {
            return null;
          }
        },
        setItem: async (name: string, value: string) => {
          try {
            const SecureStorage = require('react-native-secure-storage').default;
            await SecureStorage.setItem(name, value);
          } catch (e) {
            console.error('Failed to save to secure storage:', e);
          }
        },
        removeItem: async (name: string) => {
          try {
            const SecureStorage = require('react-native-secure-storage').default;
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
