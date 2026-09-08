/**
 * Web shim for react-native-mmkv-storage using localStorage
 */
export class MMKVLoader {
  withInstanceID(id: string) {
    return this;
  }
  withEncryption() {
    return this;
  }
  initialize() {
    return {
      getString: (key: string) => localStorage.getItem('mmkv:' + key),
      setString: (key: string, val: string) => localStorage.setItem('mmkv:' + key, val),
      getBool: (key: string) => localStorage.getItem('mmkv:' + key) === 'true',
      setBool: (key: string, val: boolean) => localStorage.setItem('mmkv:' + key, String(val)),
      removeItem: (key: string) => localStorage.removeItem('mmkv:' + key),
      clearStore: () => {
        for (let i = 0; i < localStorage.length; i++) {
          const k = localStorage.key(i);
          if (k && k.startsWith('mmkv:')) localStorage.removeItem(k);
        }
      },
    };
  }
}

export default {
  MMKVLoader,
};
