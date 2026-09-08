/**
 * Copyright (c) 2026 Bivex
 * Paywall Mobile SDK - Type Definitions
 */

export interface PaywallTheme {
  mode: 'light' | 'dark';
  accentColor: string;
  backgroundColor: string;
  surfaceColor: string;
  textColor: string;
}

export type PaywallLayout = 'centered' | 'split' | 'stacked';

export interface PaywallPlan {
  id: string;
  title: string;
  price: string;
  period: string;
  caption?: string;
  badge?: string;
  highlight?: boolean;
  productId?: string;
}

export interface PaywallHero {
  badge?: string;
  title: string;
  subtitle: string;
  socialProof?: string;
}

export interface PaywallCTA {
  primaryLabel: string;
  secondaryLabel?: string;
}

export interface PaywallFooter {
  restoreLabel: string;
  legalText: string;
}

export interface PaywallDefinition {
  id: string;
  name: string;
  platform: 'universal' | 'web' | 'mobile';
  layout: PaywallLayout;
  theme: PaywallTheme;
  hero: PaywallHero;
  features: string[];
  plans: PaywallPlan[];
  cta: PaywallCTA;
  footer: PaywallFooter;
}

export interface PaywallSDKConfig {
  appId: string;
  baseUrl?: string;
  apiKey?: string;
  debug?: boolean;
  offlineFallback?: PaywallDefinition;
}

export interface CustomerInfo {
  userId: string;
  platformUserId?: string;
  status: 'active' | 'grace_period' | 'cancelled' | 'expired' | 'none';
  planType?: string;
  expiresAt?: string;
  entitlements: Record<string, boolean>;
  hasActiveSubscription: boolean;
}

export interface PurchaseResult {
  success: boolean;
  cancelled?: boolean;
  productId?: string;
  transactionId?: string;
  customerInfo?: CustomerInfo;
  error?: string;
}

export interface RestoreResult {
  success: boolean;
  restoredCount: number;
  customerInfo?: CustomerInfo;
  error?: string;
}
