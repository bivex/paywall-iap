/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-02-28 19:34
 * Last Updated: 2026-02-28 19:34
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

export enum PaywallType {
  Premium = 'premium',
  PremiumPlus = 'premium_plus',
  Enterprise = 'enterprise',
}

export interface PaywallConfig {
  type: PaywallType;
  title: string;
  features: string[];
  priceMonthly: string;
  priceAnnual: string;
  ctaText: string;
  showTrial: boolean;
}

export const PAYWALL_CONFIGS: Record<PaywallType, PaywallConfig> = {
  [PaywallType.Premium]: {
    type: PaywallType.Premium,
    title: 'Unlock Premium',
    features: [
      'Ad-free experience',
      'Full content library',
      'Offline downloads',
      'HD quality',
    ],
    priceMonthly: '$9.99/mo',
    priceAnnual: '$79.99/year (33% savings)',
    ctaText: 'Subscribe Now',
    showTrial: false,
  },
  [PaywallType.PremiumPlus]: {
    type: PaywallType.PremiumPlus,
    title: 'Go Premium Plus',
    features: [
      'Everything in Premium',
      'Exclusive content',
      'Early access',
      'Priority support',
    ],
    priceMonthly: '$14.99/mo',
    priceAnnual: '$119.99/year (33% savings)',
    ctaText: 'Upgrade Now',
    showTrial: false,
  },
  [PaywallType.Enterprise]: {
    type: PaywallType.Enterprise,
    title: 'Enterprise Plan',
    features: [
      'All features included',
      'Custom solutions',
      'Dedicated support',
      'SLA guarantee',
    ],
    priceMonthly: 'Custom pricing',
    priceAnnual: 'Contact sales',
    ctaText: 'Contact Us',
    showTrial: false,
  },
};
