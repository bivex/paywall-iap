import {IAPService, IAPConfig} from '../../infrastructure/iap/IAPService';
import {ApiClient} from '../../infrastructure/api/ApiClient';
import {AuthService} from '../../infrastructure/api/AuthService';
import {SubscriptionService} from '../../infrastructure/api/SubscriptionService';

/**
 * Central service container for dependency injection.
 * Initialize all services here to ensure proper singleton behavior.
 */

let apiClient: ApiClient | null = null;
let authService: AuthService | null = null;
let subscriptionService: SubscriptionService | null = null;
let iapService: IAPService | null = null;

export const initializeServices = async () => {
  // Initialize API client
  apiClient = new ApiClient();

  // Initialize auth and subscription services
  authService = new AuthService(apiClient);
  subscriptionService = new SubscriptionService(apiClient);

  // Initialize IAP service with configuration
  const iapConfig: IAPConfig = {
    ios: {
      bundleId: 'com.yourapp.paywall',
      productIds: [
        'com.yourapp.premium.monthly',
        'com.yourapp.premium.yearly',
        'com.yourapp.premium_plus.monthly',
        'com.yourapp.premium_plus.yearly',
      ],
    },
    android: {
      base64PublicKey: process.env.ANDROID_PUBLIC_KEY || '',
      productIds: [
        'com.yourapp.premium.monthly',
        'com.yourapp.premium.yearly',
        'com.yourapp.premium_plus.monthly',
        'com.yourapp.premium_plus.yearly',
      ],
    },
  };

  iapService = new IAPService(iapConfig, subscriptionService);
};

export const getApiClient = (): ApiClient => {
  if (!apiClient) throw new Error('Services not initialized');
  return apiClient;
};

export const getAuthService = (): AuthService => {
  if (!authService) throw new Error('Services not initialized');
  return authService;
};

export const getSubscriptionService = (): SubscriptionService => {
  if (!subscriptionService) throw new Error('Services not initialized');
  return subscriptionService;
};

export const getIAPService = (): IAPService => {
  if (!iapService) throw new Error('Services not initialized');
  return iapService;
};

export const cleanupServices = async () => {
  if (iapService) {
    await iapService.cleanup();
  }
};
