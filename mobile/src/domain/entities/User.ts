export enum Platform {
  iOS = 'ios',
  Android = 'android',
}

export interface User {
  id: string;
  platformUserId: string;
  deviceId: string;
  platform: Platform;
  appVersion: string;
  email?: string;
  createdAt: string;
}
