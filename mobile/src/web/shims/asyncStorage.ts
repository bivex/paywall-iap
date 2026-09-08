/**
 * Web shim for @react-native-async-storage/async-storage
 */
const AsyncStorage = {
  getItem: async (key: string): Promise<string | null> => {
    return localStorage.getItem(key);
  },
  setItem: async (key: string, value: string): Promise<void> => {
    localStorage.setItem(key, value);
  },
  removeItem: async (key: string): Promise<void> => {
    localStorage.removeItem(key);
  },
  clear: async (): Promise<void> => {
    localStorage.clear();
  },
  getAllKeys: async (): Promise<string[]> => {
    const keys: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const k = localStorage.key(i);
      if (k) keys.push(k);
    }
    return keys;
  },
  multiGet: async (keys: string[]): Promise<[string, string | null][]> => {
    return keys.map((key) => [key, localStorage.getItem(key)]);
  },
  multiSet: async (keyValuePairs: [string, string][]): Promise<void> => {
    keyValuePairs.forEach(([k, v]) => localStorage.setItem(k, v));
  },
  multiRemove: async (keys: string[]): Promise<void> => {
    keys.forEach((k) => localStorage.removeItem(k));
  },
};

export default AsyncStorage;
