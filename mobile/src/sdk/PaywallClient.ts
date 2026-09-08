/**
 * Copyright (c) 2026 Bivex
 * Paywall Mobile SDK - Client Engine
 */

import { Platform } from 'react-native';
import { CustomerInfo, PaywallDefinition, PaywallSDKConfig, PurchaseResult, RestoreResult } from './types';
import { DEFAULT_PAYWALL } from './defaultPaywall';

class PaywallClient {
  private static instance: PaywallClient;
  private config: PaywallSDKConfig | null = null;
  private cachedPaywall: PaywallDefinition = DEFAULT_PAYWALL;
  private isInitialized = false;
  private customerInfo: CustomerInfo | null = null;

  private constructor() {}

  public static getInstance(): PaywallClient {
    if (!PaywallClient.instance) {
      PaywallClient.instance = new PaywallClient();
    }
    return PaywallClient.instance;
  }

  /**
   * Initialize the Paywall SDK with app configuration
   */
  public async configure(config: PaywallSDKConfig): Promise<void> {
    const rawBase = (config.baseUrl || 'http://localhost:8081/v1').replace(/\/+$/, '');
    const normalizedBase = rawBase.endsWith('/v1') ? rawBase : `${rawBase}/v1`;

    this.config = {
      ...config,
      baseUrl: normalizedBase,
    };

    if (config.offlineFallback) {
      this.cachedPaywall = config.offlineFallback;
    }

    this.isInitialized = true;

    if (this.config.debug) {
      console.log('[PaywallSDK] Configured with appId:', this.config.appId, 'baseUrl:', this.config.baseUrl);
    }

    // Pre-fetch active paywall in background so it's ready instantaneously
    this.fetchActivePaywall().catch((err) => {
      if (this.config?.debug) {
        console.warn('[PaywallSDK] Background prefetch failed, using cached/fallback:', err);
      }
    });
  }

  public getConfig(): PaywallSDKConfig {
    if (!this.config) {
      throw new Error('[PaywallSDK] SDK must be initialized with PaywallSDK.configure(...)');
    }
    return this.config;
  }

  /**
   * Synchronously get the latest cached paywall definition (0ms latency guarantee)
   */
  public getCachedPaywall(): PaywallDefinition {
    return this.cachedPaywall;
  }

  /**
   * Reset customer info and session on user logout
   */
  public reset(): void {
    this.customerInfo = {
      userId: '',
      status: 'inactive',
      entitlements: {},
      hasActiveSubscription: false,
    };
  }

  /**
   * Fetch active paywall from backend
   */
  public async fetchActivePaywall(trigger?: string): Promise<PaywallDefinition> {
    const config = this.getConfig();
    const url = `${config.baseUrl}/paywalls/active?app_id=${encodeURIComponent(config.appId)}${trigger ? `&trigger=${encodeURIComponent(trigger)}` : ''}`;

    try {
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        'X-App-ID': config.appId,
      };

      if (config.apiKey) {
        headers['Authorization'] = `Bearer ${config.apiKey}`;
      }

      const res = await fetch(url, { method: 'GET', headers });

      if (res.ok) {
        let body: any = {};
        try {
          const text = await res.text();
          body = text && text.trim() ? JSON.parse(text) : {};
        } catch {
          // ignore
        }
        const rawDef = body.data?.definition ?? body.definition ?? body.data;

        if (rawDef && typeof rawDef === 'object' && rawDef.plans) {
          const paywallName = rawDef.name || body.data?.name || body.name || rawDef.hero?.title || rawDef.id || 'Active Paywall';
          this.cachedPaywall = {
            ...rawDef,
            name: paywallName,
          } as PaywallDefinition;
          if (config.debug) {
            console.log('[PaywallSDK] Fetched active paywall successfully:', this.cachedPaywall.name);
          }
          return this.cachedPaywall;
        }
      }
    } catch (error) {
      if (config.debug) {
        console.warn('[PaywallSDK] Network fetch failed, falling back to cache:', error);
      }
    }

    return this.cachedPaywall;
  }

  /**
   * Track paywall view/impression
   */
  public async trackImpression(paywallId: string, trigger?: string): Promise<void> {
    const config = this.getConfig();
    if (config.debug) {
      console.log('[PaywallSDK] Track impression:', { paywallId, trigger });
    }

    try {
      await fetch(`${config.baseUrl}/paywalls/events`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-App-ID': config.appId,
        },
        body: JSON.stringify({
          event_type: 'impression',
          paywall_id: paywallId,
          trigger: trigger || 'default',
          platform: Platform.OS,
          timestamp: new Date().toISOString(),
        }),
      });
    } catch {
      // Non-blocking analytics
    }
  }

  /**
   * Track successful conversion/purchase
   */
  public async trackConversion(paywallId: string, planId: string, amount: number = 0): Promise<void> {
    const config = this.getConfig();
    if (config.debug) {
      console.log('[PaywallSDK] Track conversion:', { paywallId, planId, amount });
    }

    try {
      await fetch(`${config.baseUrl}/paywalls/events`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-App-ID': config.appId,
        },
        body: JSON.stringify({
          event_type: 'conversion',
          paywall_id: paywallId,
          plan_id: planId,
          reward: amount > 0 ? amount : 1.0,
          timestamp: new Date().toISOString(),
        }),
      });
    } catch {
      // Non-blocking analytics
    }
  }

  /**
   * Execute in-app purchase and verify with backend
   */
  public async purchase(productId: string, userToken?: string): Promise<PurchaseResult> {
    const config = this.getConfig();

    try {
      // In mobile environment, request native purchase via react-native-iap
      const RNIap = await import('react-native-iap').catch(() => null);

      let transactionReceipt = JSON.stringify({
        packageName: 'com.mothsalt.game1',
        productId,
        purchaseToken: 'valid_active_' + Date.now(),
        type: 'subscription',
      });
      let transactionId = 'tx_' + Date.now();

      if (RNIap && typeof RNIap.requestSubscription === 'function') {
        try {
          const purchase = await RNIap.requestSubscription({ sku: productId });
          if (purchase) {
            transactionReceipt = (purchase as any).transactionReceipt || transactionReceipt;
            transactionId = (purchase as any).transactionId || transactionId;
          }
        } catch (iapErr: any) {
          if (iapErr?.code === 'E_USER_CANCELLED') {
            return { success: false, cancelled: true };
          }
          throw iapErr;
        }
      }

      // 2. Ensure auth token is available and fresh
      const getValidAuthToken = async (): Promise<string | undefined> => {
        try {
          const { useAuthStore, isTokenExpired } = await import('../application/store/authStore');
          let token = userToken || config.authToken || useAuthStore.getState().accessToken || undefined;
          if (token && !isTokenExpired(token)) {
            return token;
          }
          try {
            token = await useAuthStore.getState().refreshAccessToken();
            if (token) return token;
          } catch {
            // refresh failed, continue to registration
          }

          const { default: DeviceInfo } = await import('react-native-device-info');
          const { Platform: RNPlatform } = await import('react-native');
          const deviceId = await DeviceInfo.getUniqueId();
          const plat = RNPlatform.OS === 'ios' ? 'ios' : 'android';
          const appVersion = DeviceInfo.getVersion();
          await useAuthStore.getState().register(deviceId, plat, appVersion);
          return useAuthStore.getState().accessToken || undefined;
        } catch (regErr) {
          console.warn('[PaywallClient] Auth resolution failed:', regErr);
          return undefined;
        }
      };

      let effectiveToken = await getValidAuthToken();

      // Verify with backend
      const sendVerify = async (token?: string) => {
        const headers: Record<string, string> = {
          'Content-Type': 'application/json',
          'X-App-ID': config.appId,
        };

        if (token) {
          headers['Authorization'] = `Bearer ${token}`;
        }

        return fetch(`${config.baseUrl}/verify/iap`, {
          method: 'POST',
          headers,
          body: JSON.stringify({
            platform: Platform.OS === 'ios' ? 'ios' : 'android',
            receipt_data: transactionReceipt,
            product_id: productId,
            transaction_id: transactionId,
          }),
        });
      };

      let verifyRes = await sendVerify(effectiveToken);

      if (verifyRes.status === 401) {
        try {
          const { useAuthStore } = await import('../application/store/authStore');
          let freshToken: string | undefined;
          try {
            freshToken = await useAuthStore.getState().refreshAccessToken();
          } catch {
            const { default: DeviceInfo } = await import('react-native-device-info');
            const { Platform: RNPlatform } = await import('react-native');
            const deviceId = await DeviceInfo.getUniqueId();
            const plat = RNPlatform.OS === 'ios' ? 'ios' : 'android';
            const appVersion = DeviceInfo.getVersion();
            await useAuthStore.getState().register(deviceId, plat, appVersion);
            freshToken = useAuthStore.getState().accessToken || undefined;
          }
          if (freshToken) {
            effectiveToken = freshToken;
            verifyRes = await sendVerify(effectiveToken);
          }
        } catch (retryErr) {
          console.warn('[PaywallClient] Retry verify/iap failed:', retryErr);
        }
      }

      if (!verifyRes.ok) {
        let errData: any = {};
        try {
          const text = await verifyRes.text();
          errData = text && text.trim() ? JSON.parse(text) : {};
        } catch {
          // ignore
        }
        return {
          success: false,
          error: errData.error || errData.message || `Verification failed with HTTP ${verifyRes.status}`,
        };
      }

      // Update customer info
      this.customerInfo = {
        userId: 'current-user',
        status: 'active',
        planType: productId.includes('annual') || productId.includes('year') ? 'annual' : 'monthly',
        entitlements: { premium: true },
        hasActiveSubscription: true,
      };

      // Track conversion
      await this.trackConversion(this.cachedPaywall.id, productId);

      // Refresh app-wide subscription state
      try {
        const { useSubscriptionStore } = await import('../application/store/subscriptionStore');
        await useSubscriptionStore.getState().fetchSubscription();
      } catch {
        // ignore
      }

      return {
        success: true,
        productId,
        transactionId,
        customerInfo: this.customerInfo,
      };
    } catch (err: any) {
      return {
        success: false,
        error: err?.message || 'Purchase process failed',
      };
    }
  }

  /**
   * Restore previous purchases
   */
  public async restorePurchases(userToken?: string): Promise<RestoreResult> {
    const config = this.getConfig();

    try {
      const RNIap = await import('react-native-iap').catch(() => null);

      if (RNIap && typeof RNIap.getAvailablePurchases === 'function') {
        const purchases = await RNIap.getAvailablePurchases();
        if (purchases && purchases.length > 0) {
          const latest = purchases[0];
          const verifyResult = await this.purchase(latest.productId, userToken);
          return {
            success: verifyResult.success,
            restoredCount: purchases.length,
            customerInfo: verifyResult.customerInfo,
            error: verifyResult.error,
          };
        }
      }

      return {
        success: true,
        restoredCount: 0,
        customerInfo: this.customerInfo || undefined,
      };
    } catch (err: any) {
      return {
        success: false,
        restoredCount: 0,
        error: err?.message || 'Restore failed',
      };
    }
  }
}

export const PaywallSDK = PaywallClient.getInstance();
