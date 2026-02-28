/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-02-28 19:36
 * Last Updated: 2026-02-28 19:36
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

// Re-export navigation types
export type {RootStackParamList, AppStackParamList, AuthStackParamList} from './Navigation';

// Navigation hook with proper typing
import {NavigationProp} from '@react-navigation/native';
import type {RootStackParamList} from './Navigation';

export type NavigationType = NavigationProp<RootStackParamList>;

// Helper to access navigation from anywhere in the app
let navigationRef: NavigationType | null = null;

export const setNavigationRef = (ref: NavigationType) => {
  navigationRef = ref;
};

export const getNavigationRef = (): NavigationType => {
  if (!navigationRef) {
    throw new Error('Navigation ref not set. Make sure Navigation component is mounted.');
  }
  return navigationRef;
};

// Navigation helpers
export const navigateToPaywall = (
  paywallType: 'premium' | 'premium_plus' | 'enterprise',
  trigger: 'subscription_expired' | 'feature_locked' | 'upgrade_prompt',
) => {
  const nav = getNavigationRef();
  nav.navigate('Paywall', {paywallType, trigger});
};

export const navigateHome = () => {
  const nav = getNavigationRef();
  nav.navigate('App');
};

export const goBack = () => {
  const nav = getNavigationRef();
  nav.goBack();
};
