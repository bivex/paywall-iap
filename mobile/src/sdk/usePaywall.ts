/**
 * Copyright (c) 2026 Bivex
 * Paywall Mobile SDK - usePaywall React Hook
 */

import { useEffect, useState } from 'react';
import { CustomerInfo, PaywallDefinition, PurchaseResult } from './types';
import { PaywallSDK } from './PaywallClient';

export interface UsePaywallResult {
  paywall: PaywallDefinition;
  loading: boolean;
  isPurchasing: boolean;
  isRestoring: boolean;
  error: string | null;
  purchase: (productId: string) => Promise<PurchaseResult>;
  restore: () => Promise<PurchaseResult>;
  refetch: () => Promise<void>;
}

export function usePaywall(trigger?: string): UsePaywallResult {
  const [paywall, setPaywall] = useState<PaywallDefinition>(PaywallSDK.getCachedPaywall());
  const [loading, setLoading] = useState(false);
  const [isPurchasing, setIsPurchasing] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchPaywall = async () => {
    setLoading(true);
    setError(null);
    try {
      const def = await PaywallSDK.fetchActivePaywall(trigger);
      setPaywall(def);
    } catch (err: any) {
      setError(err?.message || 'Failed to load paywall');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchPaywall();
  }, [trigger]);

  const purchase = async (productId: string): Promise<PurchaseResult> => {
    setIsPurchasing(true);
    setError(null);
    try {
      const result = await PaywallSDK.purchase(productId);
      if (!result.success && !result.cancelled) {
        setError(result.error || 'Purchase failed');
      }
      return result;
    } catch (err: any) {
      const msg = err?.message || 'Purchase failed';
      setError(msg);
      return { success: false, error: msg };
    } finally {
      setIsPurchasing(false);
    }
  };

  const restore = async (): Promise<PurchaseResult> => {
    setIsRestoring(true);
    setError(null);
    try {
      const result = await PaywallSDK.restorePurchases();
      if (!result.success) {
        setError(result.error || 'Restore failed');
      }
      return {
        success: result.success,
        customerInfo: result.customerInfo,
        error: result.error,
      };
    } catch (err: any) {
      const msg = err?.message || 'Restore failed';
      setError(msg);
      return { success: false, error: msg };
    } finally {
      setIsRestoring(false);
    }
  };

  return {
    paywall,
    loading,
    isPurchasing,
    isRestoring,
    error,
    purchase,
    restore,
    refetch: fetchPaywall,
  };
}
