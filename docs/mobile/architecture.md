# Mobile App Architecture

## Overview

The mobile application follows **Clean Architecture** principles with clear separation of concerns across four layers:

```
Presentation → Application → Domain ← Infrastructure
```

## Directory Structure

```
mobile/
├── src/
│   ├── domain/                    # Core business logic
│   │   └── entities/              # User, Subscription, Paywall
│   │
│   ├── application/               # Business logic & state management
│   │   ├── store/                 # Zustand stores
│   │   │   ├── authStore.ts       # Authentication state
│   │   │   ├── subscriptionStore.ts # Subscription state
│   │   │   └── iapStore.ts        # IAP state
│   │   └── services/              # Service container
│   │       └── Services.ts        # Dependency injection
│   │
│   ├── infrastructure/            # External concerns
│   │   ├── api/                   # HTTP clients
│   │   │   ├── ApiClient.ts       # Base HTTP client
│   │   │   ├── AuthService.ts     # Auth API
│   │   │   ├── SubscriptionService.ts # Subscription API
│   │   │   └── config.ts          # API configuration
│   │   ├── iap/                   # In-App Purchase
│   │   │   └── IAPService.ts      # react-native-iap wrapper
│   │   └── storage/               # Secure storage
│   │       └── SecureStorage.ts   # Keychain/Keystore wrapper
│   │
│   └── presentation/              # UI layer
│       ├── navigation/            # React Navigation
│       │   ├── Navigation.tsx     # Root navigator
│       │   └── types.ts           # Navigation types
│       └── screens/               # Screen components
│           ├── PaywallScreen.tsx
│           ├── HomeScreen.tsx
│           ├── SubscriptionScreen.tsx
│           ├── ProfileScreen.tsx
│           ├── SettingsScreen.tsx
│           ├── WelcomeScreen.tsx
│           └── RegisterScreen.tsx
│
├── App.tsx                        # Entry point
├── package.json
├── tsconfig.json
└── README.md
```

## Layer Responsibilities

### Domain Layer

**Purpose**: Enterprise business rules, completely independent of other layers.

**Entities**:
- `User` - User identity with platform-based authentication
- `Subscription` - Subscription status, plan type, access checks
- `Paywall` - Paywall configurations (Premium, PremiumPlus, Enterprise)

**Characteristics**:
- Pure TypeScript (no React Native dependencies)
- No external dependencies
- Testable in isolation

### Application Layer

**Purpose**: Application-specific business logic and state management.

**Zustand Stores**:

1. **authStore** - Authentication state
   - User session management
   - JWT token handling (access/refresh)
   - Login/register/logout actions
   - Persistent storage via AsyncStorage

2. **subscriptionStore** - Subscription state
   - Current subscription info
   - Access check results
   - Subscription refresh actions

3. **iapStore** - IAP state
   - Product catalog
   - Purchase flow management
   - Error handling
   - Loading states

**Services**:
- `Services.ts` - Central service container for dependency injection

### Infrastructure Layer

**Purpose**: External concerns and technical implementations.

**API Clients**:
- `ApiClient` - Base HTTP client with JWT management
  - Automatic token refresh
  - Error handling
  - Request/response interceptors

- `AuthService` - Authentication API
  - Device-based registration
  - Token management
  - Session persistence

- `SubscriptionService` - Subscription API
  - Fetch current subscription
  - Access verification
  - IAP receipt verification

**IAP Service**:
- `IAPService` - react-native-iap wrapper
  - Platform-specific initialization (iOS/Android)
  - Product fetching
  - Purchase flow
  - Receipt verification with backend
  - Transaction management

**Storage**:
- `SecureStorage` - Secure storage wrapper
  - iOS Keychain
  - Android Keystore
  - Fallback to AsyncStorage (non-sensitive data)

### Presentation Layer

**Purpose**: UI components and user interaction.

**Navigation**:
- Root Stack: Auth flow → App tabs
- Auth Stack: Welcome → Register
- App Tabs: Home, Subscription, Profile, Settings

**Screens**:

1. **WelcomeScreen** - Onboarding entry point
   - Auto-registration on mount
   - Navigate to paywall or home based on subscription

2. **RegisterScreen** - Device registration
   - Platform detection
   - Device ID collection
   - Automatic registration

3. **PaywallScreen** - Subscription purchase
   - Dynamic paywall configs (Premium, PremiumPlus, Enterprise)
   - IAP purchase flow
   - Restore purchases
   - Terms/privacy links

4. **HomeScreen** - Main content
   - Access check display
   - Subscription status
   - Navigate to paywall if needed

5. **SubscriptionScreen** - Subscription management
   - Current subscription details
   - Plan change options
   - Restore purchases

6. **ProfileScreen** - User profile
   - User info display
   - Account details

7. **SettingsScreen** - App settings
   - Logout
   - App preferences

## Data Flow

### Authentication Flow

```
1. App.tsx initializes Services
2. AuthService.register() called with device info
3. Backend returns JWT tokens (access + refresh)
4. Tokens stored in SecureStorage
5. authStore updated with user + tokens
6. Navigation routes to Home or Paywall based on subscription
```

### Purchase Flow

```
1. User taps subscription plan on PaywallScreen
2. iapStore.purchase(productId) called
3. IAPService.requestPurchase() → react-native-iap
4. Platform purchase dialog shown
5. On success: purchaseUpdatedListener fires
6. IAPService.verifyWithBackend() → SubscriptionService.verifyIAP()
7. Backend validates receipt with Apple/Google
8. Backend creates/updates subscription
9. IAPService.finishPurchase() completes transaction
10. subscriptionStore refreshed
11. Navigation routes to Home
```

### Access Check Flow

```
1. HomeScreen calls subscriptionStore.checkAccess()
2. SubscriptionService.getCurrentSubscription() → backend
3. Backend checks subscription status + expiration
4. Returns AccessCheck { hasAccess: boolean, reason: string }
5. subscriptionStore updated
6. UI renders accordingly
```

## State Management

### Zustand Pattern

```typescript
interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  login: (email, password) => Promise<void>;
  logout: () => Promise<void>;
  refreshAccessToken: () => Promise<void>;
  loadStoredTokens: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // State
      user: null,
      accessToken: null,
      refreshToken: null,
      isLoading: false,
      error: null,
      
      // Actions
      login: async (email, password) => {
        // Implementation
      },
      // ...
    }),
    {
      name: 'auth-storage',
      storage: createJSONStorage(() => AsyncStorage),
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    }
  )
);
```

## Dependency Injection

### Service Container

```typescript
class Services {
  private static instance: Services;
  
  public readonly apiClient: ApiClient;
  public readonly authService: AuthService;
  public readonly subscriptionService: SubscriptionService;
  public readonly iapService: IAPService;
  
  private constructor() {
    this.apiClient = new ApiClient(API_CONFIG);
    this.authService = new AuthService(this.apiClient);
    this.subscriptionService = new SubscriptionService(this.apiClient);
    this.iapService = new IAPService(IAP_CONFIG, this.subscriptionService);
  }
  
  public static getInstance(): Services {
    if (!Services.instance) {
      Services.instance = new Services();
    }
    return Services.instance;
  }
}
```

## Security

### Token Management

- **Access Token**: Short-lived (15 minutes), stored in memory
- **Refresh Token**: Long-lived (30 days), stored in SecureStorage
- **Automatic Refresh**: ApiClient intercepts 401 responses and refreshes

### Secure Storage

- **iOS**: Keychain with `kSecAttrAccessibleWhenUnlocked`
- **Android**: EncryptedSharedPreferences with Keystore
- **Keys Stored**: Refresh token, device ID

### Receipt Verification

All IAP receipts are verified with the backend:
1. Mobile sends receipt data to backend
2. Backend validates with Apple/Google
3. Backend creates subscription record
4. Mobile receives confirmation

## Testing Strategy

### Unit Tests

- Domain entities (pure business logic)
- Store actions (state transformations)
- Service methods (mocked dependencies)

### Integration Tests

- API client with mock server
- IAP flow with test environment
- Navigation flow

### E2E Tests

- Complete purchase flow (TestFlight/Play Console)
- Auth flow with test backend
- Subscription management

## Platform-Specific Considerations

### iOS

- **IAP**: Uses StoreKit 2 via react-native-iap
- **Storage**: Keychain with accessibility attributes
- **Navigation**: Native stack navigator
- **Receipt**: Base64-encoded transactionReceipt

### Android

- **IAP**: Uses Google Play Billing Library v5
- **Storage**: EncryptedSharedPreferences
- **Navigation**: Native stack navigator
- **Receipt**: JSON purchaseToken data

## Performance Optimizations

1. **Lazy Loading**: Screens loaded on demand
2. **Memoization**: React.memo for static screens
3. **Caching**: Zustand persist for offline access
4. **Batching**: Multiple state updates batched
5. **Image Optimization**: Cached images with react-native-fast-image

## Error Handling

### Network Errors

```typescript
try {
  await apiClient.post('/subscription/verify', data);
} catch (error) {
  if (error instanceof NetworkError) {
    // Show offline message
  } else if (error instanceof ApiError) {
    // Show API error message
  } else {
    // Log unexpected error
  }
}
```

### IAP Errors

```typescript
purchaseErrorListener((error) => {
  console.error('[IAP] Purchase error:', error);
  // Handle specific error codes
  switch (error.code) {
    case 'E_USER_CANCELLED':
      // User cancelled - no action needed
      break;
    case 'E_SERVICE_UNAVAILABLE':
      // Store service unavailable - retry later
      break;
    default:
      // Show error message
  }
});
```

## Future Enhancements

1. **Offline Mode**: Cache subscription status for offline access
2. **Promo Codes**: Support for promotional offers
3. **Intro Trials**: Handle introductory trial periods
4. **Family Sharing**: iOS Family Sharing support
5. **Subscription Offers**: Display promotional offers
6. **A/B Testing**: Remote config for paywall variants
7. **Analytics**: Purchase funnel analytics
8. **Crash Reporting**: Sentry integration

## Related Documents

- [API Specification](../api/openapi.yaml)
- [Database Schema](../database/schema-erd.md)
- [IAP Service Implementation](../../mobile/src/infrastructure/iap/IAPService.ts)
- [Store Implementation](../../mobile/src/application/store/)
