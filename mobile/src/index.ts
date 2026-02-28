// Domain entities
export * from './domain/entities/User';
export * from './domain/entities/Subscription';
export * from './domain/entities/Paywall';

// Stores
export {useAuthStore} from './application/store/authStore';
export {useSubscriptionStore} from './application/store/subscriptionStore';
export {useIAPStore, useProducts, useIsPurchasing, useIAPError} from './application/store/iapStore';

// Services
export {initializeServices, getApiClient, getAuthService, getSubscriptionService, getIAPService, cleanupServices} from './application/services/Services';

// Infrastructure
export {IAPService, type IAPProduct, type IAPConfig} from './infrastructure/iap/IAPService';
export {ApiClient} from './infrastructure/api/ApiClient';
export {AuthService, type RegisterRequest, type RegisterResponse} from './infrastructure/api/AuthService';
export {SubscriptionService, type VerifyIAPRequest, type VerifyIAPResponse} from './infrastructure/api/SubscriptionService';
export {SecureStorage} from './infrastructure/storage/SecureStorage';

// Navigation
export {Navigation} from './presentation/navigation/Navigation';
export type {RootStackParamList, AppStackParamList, AuthStackParamList} from './presentation/navigation/types';
export {navigateToPaywall, navigateHome, goBack} from './presentation/navigation/types';
