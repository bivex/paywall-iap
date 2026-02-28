# Paywall IAP - Mobile App

React Native mobile application with in-app purchase (IAP) support for iOS and Android.

## Features

- **Authentication**: Device-based registration with JWT token management
- **In-App Purchases**: Full IAP integration via `react-native-iap`
- **Subscription Management**: View and manage subscriptions
- **Paywall System**: Dynamic paywall with multiple tiers
- **State Management**: Zustand for global state
- **Navigation**: React Navigation with authenticated and unauthenticated flows

## Tech Stack

- **Framework**: React Native 0.73.6
- **Language**: TypeScript
- **State Management**: Zustand 4.5.0
- **Navigation**: React Navigation 6.x
- **IAP**: react-native-iap 12.10.5
- **HTTP Client**: Axios 1.6.7
- **Secure Storage**: react-native-secure-storage

## Project Structure

```
mobile/
├── src/
│   ├── application/          # Application layer
│   │   ├── services/         # Service initialization
│   │   └── store/            # Zustand stores (auth, subscription, IAP)
│   ├── domain/               # Domain layer
│   │   └── entities/         # Domain entities (User, Subscription, Paywall)
│   ├── infrastructure/       # Infrastructure layer
│   │   ├── api/              # API clients (Auth, Subscription)
│   │   ├── iap/              # IAP service
│   │   └── storage/          # Secure storage wrapper
│   └── presentation/         # Presentation layer
│       ├── navigation/       # Navigation setup
│       └── screens/          # Screen components
├── App.tsx                   # App entry point
├── package.json
└── tsconfig.json
```

## Getting Started

### Prerequisites

- Node.js 18+
- React Native CLI
- iOS: Xcode 15+ (for iOS development)
- Android: Android Studio with SDK 33+ (for Android development)

### Installation

```bash
cd mobile
npm install
```

### Configuration

Create a `.env` file in the project root:

```env
API_BASE_URL=http://localhost:8080/v1
ANDROID_PUBLIC_KEY=<your-google-play-base64-key>
```

### Running the App

**iOS:**
```bash
npx pod-install
npm run ios
```

**Android:**
```bash
npm run android
```

## IAP Configuration

### iOS

1. Add your product IDs in `src/application/services/Services.ts`:
```typescript
ios: {
  bundleId: 'com.yourapp.paywall',
  productIds: [
    'com.yourapp.premium.monthly',
    'com.yourapp.premium.yearly',
  ],
}
```

2. Configure products in App Store Connect

### Android

1. Add your Google Play public key to `.env`:
```
ANDROID_PUBLIC_KEY=<base64-encoded-public-key>
```

2. Add product IDs in `src/application/services/Services.ts`

3. Configure products in Google Play Console

## Architecture

This app follows **Clean Architecture** principles with clear separation of concerns:

- **Domain**: Core business entities (no dependencies)
- **Application**: Use cases, state management, service coordination
- **Infrastructure**: External integrations (API, IAP, storage)
- **Presentation**: UI components and navigation

## Key Services

### IAPService

Handles in-app purchases:
```typescript
const iapService = getIAPService();
await iapService.initialize();
const products = await iapService.getProducts();
const purchase = await iapService.purchase(productId);
await iapService.finishPurchase(purchase);
```

### AuthService

Manages authentication:
```typescript
const authService = getAuthService();
await authService.register(platform, deviceId, appVersion);
const {accessToken, refreshToken} = await authService.getStoredTokens();
```

### SubscriptionService

Manages subscriptions:
```typescript
const subscriptionService = getSubscriptionService();
const subscription = await subscriptionService.getSubscription();
await subscriptionService.verifyIAP(platform, receiptData, productId);
```

## State Management

### Auth Store
```typescript
const {user, isAuthenticated, register, logout} = useAuthStore();
```

### Subscription Store
```typescript
const {subscription, fetchSubscription, checkAccess} = useSubscriptionStore();
```

### IAP Store
```typescript
const {products, purchase, finishPurchase} = useIAPStore();
```

## Navigation

Programmatic navigation:
```typescript
import {navigateToPaywall, navigateHome} from '@mobile';

// Navigate to paywall
navigateToPaywall('premium', 'feature_locked');

// Navigate home
navigateHome();
```

## License

MIT

## Author

Bivex - support@b-b.top
