import {Platform} from 'react-native';
import SecureStorageDefault from 'react-native-secure-storage';

// Secure storage wrapper for cross-platform compatibility
export const SecureStorage = {
  async setItem(key: string, value: string): Promise<void> {
    try {
      await SecureStorageDefault.setItem(key, value);
    } catch (error) {
      console.error(`[SecureStorage] Failed to set ${key}:`, error);
      throw error;
    }
  },

  async getItem(key: string): Promise<string | null> {
    try {
      return await SecureStorageDefault.getItem(key);
    } catch (error) {
      console.error(`[SecureStorage] Failed to get ${key}:`, error);
      return null;
    }
  },

  async removeItem(key: string): Promise<void> {
    try {
      await SecureStorageDefault.removeItem(key);
    } catch (error) {
      console.error(`[SecureStorage] Failed to remove ${key}:`, error);
      throw error;
    }
  },

  async clear(): Promise<void> {
    try {
      // Clear all auth-related keys
      const keys = ['access_token', 'refresh_token', 'user_id', 'device_id'];
      for (const key of keys) {
        await SecureStorageDefault.removeItem(key);
      }
    } catch (error) {
      console.error('[SecureStorage] Failed to clear:', error);
      throw error;
    }
  },
};
