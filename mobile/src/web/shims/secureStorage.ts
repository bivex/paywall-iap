/**
 * Web shim for react-native-secure-storage using localStorage
 */
const PREFIX = 'secure_storage:';

export const SecureStorage = {
  async setItem(key: string, value: string): Promise<void> {
    try {
      localStorage.setItem(PREFIX + key, value);
    } catch (e) {
      console.warn('[SecureStorage Web] setItem error:', e);
    }
  },

  async getItem(key: string): Promise<string | null> {
    try {
      return localStorage.getItem(PREFIX + key);
    } catch {
      return null;
    }
  },

  async removeItem(key: string): Promise<void> {
    try {
      localStorage.removeItem(PREFIX + key);
    } catch (e) {
      console.warn('[SecureStorage Web] removeItem error:', e);
    }
  },

  async clear(): Promise<void> {
    try {
      const keysToRemove: string[] = [];
      for (let i = 0; i < localStorage.length; i++) {
        const k = localStorage.key(i);
        if (k && k.startsWith(PREFIX)) {
          keysToRemove.push(k);
        }
      }
      keysToRemove.forEach((k) => localStorage.removeItem(k));
    } catch (e) {
      console.warn('[SecureStorage Web] clear error:', e);
    }
  },
};

export default SecureStorage;
