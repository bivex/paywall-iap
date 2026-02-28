/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-02-28 19:33
 * Last Updated: 2026-02-28 19:33
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

import {Platform} from 'react-native';
import {
  Product,
  purchaseErrorListener,
  purchaseUpdatedListener,
  SubscriptionPurchase,
  initConnection,
  endConnection,
  getProducts,
  requestSubscription,
  finishTransaction,
  SubscriptionAndroid,
  flushFailedPurchasesIOS,
  getSubscriptionPurchases,
  getSubscriptions,
} from 'react-native-iap';
import {SubscriptionService} from '../api/SubscriptionService';

export interface IAPProduct {
  productId: string;
  price: string;
  currency: string;
  localizedTitle: string;
  localizedDescription: string;
}

export interface IAPConfig {
  ios: {
    bundleId: string;
    productIds: string[];
  };
  android: {
    base64PublicKey: string;
    productIds: string[];
  };
}

export class IAPService {
  private products: Map<string, Product> = new Map();
  private subscriptions: Map<string, SubscriptionAndroid> = new Map();
  private initialized = false;

  constructor(
    private config: IAPConfig,
    private subscriptionService: SubscriptionService,
  ) {}

  async initialize(): Promise<void> {
    if (this.initialized) {
      return;
    }

    try {
      await initConnection();

      // Set up listeners
      purchaseErrorListener((error: Error) => {
        console.error('[IAP] Purchase error:', error);
      });

      purchaseUpdatedListener((purchase: SubscriptionPurchase) => {
        console.log('[IAP] Purchase updated:', purchase);
        this.processPurchaseUpdate(purchase);
      });

      // Get products/skus based on platform
      if (Platform.OS === 'ios') {
        const products = await getProducts(this.config.ios.productIds);
        for (const product of products) {
          this.products.set(product.productId, product);
        }
      } else {
        const subscriptions = await getSubscriptions(this.config.android.productIds);
        for (const subscription of subscriptions) {
          this.products.set(subscription.productId, subscription);
        }
      }

      // Get existing subscriptions (Android only)
      if (Platform.OS === 'android') {
        const purchases = await getSubscriptionPurchases();
        for (const purchase of purchases) {
          this.subscriptions.set(purchase.productId, purchase as SubscriptionAndroid);
        }
      }

      this.initialized = true;
      console.log('[IAP] Initialized successfully');
    } catch (error) {
      console.error('[IAP] Failed to initialize:', error);
      throw error;
    }
  }

  async getProducts(): Promise<IAPProduct[]> {
    await this.ensureInitialized();

    return Array.from(this.products.values()).map((p) => ({
      productId: p.productId,
      price: p.price,
      currency: p.currency,
      localizedTitle: p.localizedTitle,
      localizedDescription: p.localizedDescription,
    }));
  }

  async purchase(productId: string): Promise<SubscriptionPurchase> {
    await this.ensureInitialized();

    const product = this.products.get(productId);
    if (!product) {
      throw new Error(`Product not found: ${productId}`);
    }

    try {
      const purchase = await requestSubscription(productId);
      return purchase;
    } catch (error) {
      console.error('[IAP] Purchase failed:', error);
      throw error;
    }
  }

  async finishPurchase(purchase: SubscriptionPurchase): Promise<void> {
    try {
      // Verify receipt with backend first
      await this.verifyWithBackend(purchase);

      // Then finish the transaction
      await finishTransaction({purchase}, true);

      console.log('[IAP] Purchase finished successfully');
    } catch (error) {
      console.error('[IAP] Failed to finish purchase:', error);
      throw error;
    }
  }

  async getCurrentSubscription(): Promise<SubscriptionAndroid | null> {
    await this.ensureInitialized();

    // For Android, return cached subscription
    if (Platform.OS === 'android' && this.subscriptions.size > 0) {
      return Array.from(this.subscriptions.values())[0];
    }

    // For iOS, we need to fetch current subscription via API
    return null;
  }

  async restorePurchases(): Promise<void> {
    try {
      if (Platform.OS === 'ios') {
        // iOS: get current purchases
        const purchases = await getSubscriptionPurchases();
        for (const purchase of purchases) {
          await this.verifyWithBackend(purchase);
        }
      } else {
        // Android: flush failed purchases
        await flushFailedPurchasesIOS();
      }
      console.log('[IAP] Purchases restored successfully');
    } catch (error) {
      console.error('[IAP] Failed to restore purchases:', error);
      throw error;
    }
  }

  async cleanup(): Promise<void> {
    try {
      await endConnection();
      this.initialized = false;
      console.log('[IAP] Cleaned up successfully');
    } catch (error) {
      console.error('[IAP] Failed to cleanup:', error);
    }
  }

  private async ensureInitialized(): Promise<void> {
    if (!this.initialized) {
      await this.initialize();
    }
  }

  private async processPurchaseUpdate(purchase: SubscriptionPurchase): Promise<void> {
    console.log('[IAP] Processing purchase update:', purchase);

    // Store subscription info for Android
    if (Platform.OS === 'android') {
      this.subscriptions.set(purchase.productId, purchase as SubscriptionAndroid);
    }

    // Verify with backend
    await this.verifyWithBackend(purchase);
  }

  private async verifyWithBackend(purchase: SubscriptionPurchase): Promise<void> {
    try {
      const platform = Platform.OS === 'ios' ? 'ios' : 'android';
      const receiptData = purchase.receiptData || '';

      await this.subscriptionService.verifyIAP(
        platform,
        receiptData,
        purchase.productId,
      );

      console.log('[IAP] Purchase verified with backend:', purchase.productId);
    } catch (error) {
      console.error('[IAP] Backend verification failed:', error);
      throw error;
    }
  }
}

export default IAPService;
