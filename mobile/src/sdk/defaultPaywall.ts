/**
 * Copyright (c) 2026 Bivex
 * Paywall Mobile SDK - Default Offline Fallback Template
 */

import { PaywallDefinition } from './types';

export const DEFAULT_PAYWALL: PaywallDefinition = {
  id: 'default-mobile-paywall',
  name: 'Default Premium Paywall',
  platform: 'mobile',
  layout: 'stacked',
  theme: {
    mode: 'dark',
    accentColor: '#6366F1', // Indigo
    backgroundColor: '#0F172A', // Slate 900
    surfaceColor: '#1E293B', // Slate 800
    textColor: '#F8FAFC', // Slate 50
  },
  hero: {
    badge: 'Limited Time Offer',
    title: 'Unlock Full Access',
    subtitle: 'Get unlimited access to all premium features, no ads, and priority support.',
    socialProof: 'Joined by 50,000+ active subscribers',
  },
  features: [
    'Unlimited access to all content & tools',
    'Offline mode & fast cloud sync',
    'Ad-free experience',
    'VIP 24/7 priority support',
  ],
  plans: [
    {
      id: 'monthly',
      title: 'Monthly Access',
      price: '$9.99',
      period: '/month',
      caption: 'Cancel anytime',
      badge: '',
      highlight: false,
      productId: 'com.mothsalt.game1.monthly',
    },
    {
      id: 'annual',
      title: 'Annual Plan',
      price: '$59.99',
      period: '/year',
      caption: '$4.99/mo (50% savings)',
      badge: 'Best Value',
      highlight: true,
      productId: 'com.mothsalt.game1.annual',
    },
  ],
  cta: {
    primaryLabel: 'Continue',
    secondaryLabel: 'Not now',
  },
  footer: {
    restoreLabel: 'Restore Purchases',
    legalText: 'Auto-renews periodically. Manage or cancel subscription anytime in your App Store / Google Play account settings.',
  },
};
