/**
 * Web shim for react-native-device-info
 */
export const DeviceInfo = {
  getUniqueId: async (): Promise<string> => {
    let id = localStorage.getItem('__web_device_id');
    if (!id) {
      id = 'web_' + Math.random().toString(36).substring(2, 15);
      localStorage.setItem('__web_device_id', id);
    }
    return id;
  },
  getVersion: (): string => '1.0.0-web',
  getBuildNumber: (): string => '1',
  getBundleId: (): string => 'com.mothsalt.game1.web',
  getModel: (): string => 'Web Browser',
  getSystemName: (): string => 'Web',
  getSystemVersion: (): string => (typeof navigator !== 'undefined' ? navigator.userAgent : 'Web'),
  isTablet: (): boolean => false,
};

export default DeviceInfo;
