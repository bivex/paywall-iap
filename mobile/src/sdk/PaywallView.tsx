/**
 * Copyright (c) 2026 Bivex
 * Paywall Mobile SDK - Server-Driven UI Component
 */

import React, { useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  Dimensions,
  Platform,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { PaywallDefinition, PaywallPlan, PurchaseResult } from './types';
import { PaywallSDK } from './PaywallClient';

const { width: SCREEN_WIDTH } = Dimensions.get('window');

export interface PaywallViewProps {
  definition?: PaywallDefinition;
  trigger?: string;
  onDismiss?: () => void;
  onPurchaseSuccess?: (result: PurchaseResult) => void;
  onError?: (error: string) => void;
}

export const PaywallView: React.FC<PaywallViewProps> = ({
  definition: initialDefinition,
  trigger,
  onDismiss,
  onPurchaseSuccess,
  onError,
}) => {
  const [paywall, setPaywall] = useState<PaywallDefinition>(
    initialDefinition || PaywallSDK.getCachedPaywall()
  );
  const [selectedPlanId, setSelectedPlanId] = useState<string>(
    paywall.plans.find((p) => p.highlight)?.id || paywall.plans[0]?.id || ''
  );
  const [isPurchasing, setIsPurchasing] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const [feedback, setFeedback] = useState<{
    type: 'success' | 'error' | 'info';
    title: string;
    message: string;
  } | null>(null);

  // Fetch active paywall on mount if not provided explicitly
  useEffect(() => {
    if (!initialDefinition) {
      PaywallSDK.fetchActivePaywall(trigger).then((fetched) => {
        if (fetched && fetched.plans?.length > 0) {
          setPaywall(fetched);
          setSelectedPlanId((current) => {
            const hasCurrent = fetched.plans.some((p) => p.id === current);
            if (hasCurrent) return current;
            const highlighted = fetched.plans.find((p) => p.highlight) || fetched.plans[0];
            return highlighted.id;
          });
        }
      });
    }

    PaywallSDK.trackImpression(paywall.id, trigger);
  }, [initialDefinition, trigger]);

  const { theme, hero, features, plans, cta, footer, layout } = paywall;
  const selectedPlan = plans.find((p) => p.id === selectedPlanId) || plans.find((p) => p.highlight) || plans[0];

  const handlePurchase = async () => {
    if (!selectedPlan) return;
    setIsPurchasing(true);
    setFeedback(null);

    try {
      const productId = selectedPlan.productId || selectedPlan.id;
      const result = await PaywallSDK.purchase(productId);

      if (result.success) {
        setFeedback({
          type: 'success',
          title: 'Purchase Successful!',
          message: 'Your subscription is now active. Enjoy premium features!',
        });
        setTimeout(() => {
          onPurchaseSuccess?.(result);
        }, 1200);
      } else if (!result.cancelled) {
        const errorMsg = result.error || 'Payment failed';
        setFeedback({
          type: 'error',
          title: 'Purchase Failed',
          message: errorMsg,
        });
        onError?.(errorMsg);
      }
    } catch (err: any) {
      const msg = err?.message || 'Payment could not be completed';
      setFeedback({
        type: 'error',
        title: 'Error',
        message: msg,
      });
      onError?.(msg);
    } finally {
      setIsPurchasing(false);
    }
  };

  const handleRestore = async () => {
    setIsRestoring(true);
    setFeedback(null);
    try {
      const result = await PaywallSDK.restorePurchases();
      if (result.success && (result.restoredCount ?? 0) > 0) {
        setFeedback({
          type: 'success',
          title: 'Purchases Restored!',
          message: `Successfully restored ${result.restoredCount} subscription(s). Refreshing access...`,
        });
        setTimeout(() => {
          onPurchaseSuccess?.({ success: true, customerInfo: result.customerInfo });
        }, 1400);
      } else if (result.success) {
        setFeedback({
          type: 'info',
          title: 'No Active Purchases',
          message: 'No previous active purchases were found to restore for this account.',
        });
      } else {
        setFeedback({
          type: 'error',
          title: 'Restore Failed',
          message: result.error || 'Could not restore purchases.',
        });
      }
    } catch (err: any) {
      setFeedback({
        type: 'error',
        title: 'Restore Error',
        message: err?.message || 'Failed to communicate with restore service',
      });
    } finally {
      setIsRestoring(false);
    }
  };

  return (
    <SafeAreaView style={[styles.container, { backgroundColor: theme.backgroundColor }]}>
      {/* Top Bar / Close Button */}
      <View style={styles.topBar}>
        <View style={{ flex: 1 }} />
        {onDismiss && (
          <TouchableOpacity
            style={[styles.closeButton, { backgroundColor: theme.surfaceColor }]}
            onPress={onDismiss}
            hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
          >
            <Text style={[styles.closeText, { color: theme.textColor }]}>✕</Text>
          </TouchableOpacity>
        )}
      </View>

      {/* Inline Feedback Banner */}
      {feedback && (
        <View
          style={[
            styles.feedbackBanner,
            feedback.type === 'success' && styles.feedbackBannerSuccess,
            feedback.type === 'error' && styles.feedbackBannerError,
            feedback.type === 'info' && styles.feedbackBannerInfo,
          ]}
        >
          <Text style={styles.feedbackIcon}>
            {feedback.type === 'success' ? '✅' : feedback.type === 'error' ? '❌' : 'ℹ️'}
          </Text>
          <View style={styles.feedbackTextCol}>
            <Text style={styles.feedbackTitle}>{feedback.title}</Text>
            <Text style={styles.feedbackMessage}>{feedback.message}</Text>
          </View>
          <TouchableOpacity
            style={styles.feedbackCloseBtn}
            onPress={() => setFeedback(null)}
            hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
          >
            <Text style={styles.feedbackCloseText}>✕</Text>
          </TouchableOpacity>
        </View>
      )}

      <ScrollView
        contentContainerStyle={[
          styles.scrollContent,
          layout === 'centered' && styles.centeredLayout,
        ]}
        showsVerticalScrollIndicator={false}
      >
        {/* Hero Section */}
        <View style={styles.heroSection}>
          {hero.badge ? (
            <View style={[styles.badge, { backgroundColor: `${theme.accentColor}25`, borderColor: theme.accentColor }]}>
              <Text style={[styles.badgeText, { color: theme.accentColor }]}>
                {hero.badge.toUpperCase()}
              </Text>
            </View>
          ) : null}

          <Text style={[styles.title, { color: theme.textColor }]}>{hero.title}</Text>
          <Text style={[styles.subtitle, { color: `${theme.textColor}B3` }]}>{hero.subtitle}</Text>

          {hero.socialProof ? (
            <View style={styles.socialProofContainer}>
              <Text style={styles.starText}>★★★★★</Text>
              <Text style={[styles.socialProofText, { color: `${theme.textColor}99` }]}>
                {hero.socialProof}
              </Text>
            </View>
          ) : null}
        </View>

        {/* Benefits List */}
        <View style={[styles.featuresSection, { backgroundColor: theme.surfaceColor }]}>
          {features.map((feature: string, index: number) => (
            <View key={index} style={styles.featureRow}>
              <View style={[styles.checkCircle, { backgroundColor: theme.accentColor }]}>
                <Text style={styles.checkIcon}>✓</Text>
              </View>
              <Text style={[styles.featureText, { color: theme.textColor }]}>{feature}</Text>
            </View>
          ))}
        </View>

        {/* Plan Selection Cards */}
        <View style={styles.plansSection}>
          {plans.map((plan: PaywallPlan) => {
            const isSelected = plan.id === selectedPlanId;
            return (
              <TouchableOpacity
                key={plan.id}
                activeOpacity={0.8}
                onPress={() => setSelectedPlanId(plan.id)}
                style={[
                  styles.planCard,
                  {
                    backgroundColor: theme.surfaceColor,
                    borderColor: isSelected ? theme.accentColor : `${theme.textColor}1A`,
                    borderWidth: isSelected ? 2 : 1,
                    ...(Platform.OS === 'web' ? { cursor: 'pointer' } : {}),
                  } as any,
                ]}
              >
                {plan.badge ? (
                  <View style={[styles.planBadge, { backgroundColor: theme.accentColor }]}>
                    <Text style={styles.planBadgeText}>{plan.badge}</Text>
                  </View>
                ) : null}

                <View style={styles.planHeader}>
                  <View style={styles.planTitleContainer}>
                    <View
                      style={[
                        styles.radioOuter,
                        { borderColor: isSelected ? theme.accentColor : `${theme.textColor}40` },
                      ]}
                    >
                      {isSelected && (
                        <View style={[styles.radioInner, { backgroundColor: theme.accentColor }]} />
                      )}
                    </View>
                    <Text style={[styles.planTitle, { color: theme.textColor }]}>{plan.title}</Text>
                  </View>

                  <View style={styles.priceContainer}>
                    <Text style={[styles.planPrice, { color: theme.textColor }]}>
                      {plan.price}
                    </Text>
                    <Text style={[styles.planPeriod, { color: `${theme.textColor}80` }]}>
                      {plan.period}
                    </Text>
                  </View>
                </View>

                {plan.caption ? (
                  <Text style={[styles.planCaption, { color: `${theme.textColor}99` }]}>
                    {plan.caption}
                  </Text>
                ) : null}
              </TouchableOpacity>
            );
          })}
        </View>

        {/* Primary CTA Button */}
        <TouchableOpacity
          activeOpacity={0.85}
          disabled={isPurchasing || isRestoring}
          onPress={handlePurchase}
          style={[
            styles.ctaButton,
            { backgroundColor: theme.accentColor },
            isPurchasing && { opacity: 0.7 },
            (Platform.OS === 'web' ? { cursor: 'pointer' } : {}),
          ] as any}
        >
          {isPurchasing ? (
            <ActivityIndicator color="#FFFFFF" />
          ) : (
            <Text style={styles.ctaText}>
              {(cta as any).primaryLabel || (cta as any).text || 'Continue'}
            </Text>
          )}
        </TouchableOpacity>

        {(cta as any).subtext ? (
          <Text style={[styles.ctaSubtext, { color: `${theme.textColor}99` }]}>
            {(cta as any).subtext}
          </Text>
        ) : null}

        {/* Secondary Dismiss Action (if configured) */}
        {cta.secondaryLabel && onDismiss ? (
          <TouchableOpacity onPress={onDismiss} style={styles.secondaryButton}>
            <Text style={[styles.secondaryText, { color: `${theme.textColor}80` }]}>
              {cta.secondaryLabel}
            </Text>
          </TouchableOpacity>
        ) : null}

        {/* Footer & Restore Purchases */}
        <View style={styles.footerSection}>
          <TouchableOpacity
            disabled={isRestoring}
            onPress={handleRestore}
            style={styles.restoreButton}
          >
            {isRestoring ? (
              <ActivityIndicator size="small" color={theme.accentColor} />
            ) : (
              <Text style={[styles.restoreText, { color: `${theme.textColor}B3` }]}>
                {(footer as any).restoreLabel || (footer as any).restoreText || 'Restore Purchases'}
              </Text>
            )}
          </TouchableOpacity>

          {(footer as any).legalText ? (
            <Text style={[styles.legalText, { color: `${theme.textColor}66` }]}>
              {(footer as any).legalText}
            </Text>
          ) : null}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  topBar: {
    flexDirection: 'row',
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 4,
    zIndex: 10,
  },
  closeButton: {
    width: 32,
    height: 32,
    borderRadius: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  closeText: {
    fontSize: 14,
    fontWeight: 'bold',
  },
  scrollContent: {
    paddingHorizontal: 20,
    paddingBottom: 36,
  },
  centeredLayout: {
    alignItems: 'center',
  },
  heroSection: {
    alignItems: 'center',
    marginVertical: 16,
  },
  badge: {
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 12,
    borderWidth: 1,
    marginBottom: 12,
  },
  badgeText: {
    fontSize: 11,
    fontWeight: '700',
    letterSpacing: 0.5,
  },
  title: {
    fontSize: 26,
    fontWeight: '800',
    textAlign: 'center',
    marginBottom: 8,
    lineHeight: 32,
  },
  subtitle: {
    fontSize: 15,
    textAlign: 'center',
    lineHeight: 22,
    maxWidth: SCREEN_WIDTH * 0.85,
  },
  socialProofContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    marginTop: 10,
  },
  starText: {
    color: '#F59E0B',
    fontSize: 14,
  },
  socialProofText: {
    fontSize: 12,
    fontWeight: '500',
  },
  featuresSection: {
    borderRadius: 16,
    padding: 16,
    marginVertical: 14,
  },
  featureRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 6,
  },
  checkCircle: {
    width: 20,
    height: 20,
    borderRadius: 10,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 12,
  },
  checkIcon: {
    color: '#FFFFFF',
    fontSize: 12,
    fontWeight: 'bold',
  },
  featureText: {
    fontSize: 14,
    fontWeight: '500',
    flex: 1,
  },
  plansSection: {
    marginVertical: 10,
    gap: 12,
  },
  planCard: {
    borderRadius: 16,
    padding: 16,
    position: 'relative',
  },
  planBadge: {
    position: 'absolute',
    top: -10,
    right: 16,
    paddingHorizontal: 10,
    paddingVertical: 2,
    borderRadius: 10,
  },
  planBadgeText: {
    color: '#FFFFFF',
    fontSize: 11,
    fontWeight: '700',
  },
  planHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  planTitleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  radioOuter: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  radioInner: {
    width: 10,
    height: 10,
    borderRadius: 5,
  },
  planTitle: {
    fontSize: 16,
    fontWeight: '700',
  },
  priceContainer: {
    alignItems: 'flex-end',
  },
  planPrice: {
    fontSize: 18,
    fontWeight: '800',
  },
  planPeriod: {
    fontSize: 11,
  },
  planCaption: {
    fontSize: 12,
    marginTop: 6,
    marginLeft: 30,
  },
  ctaButton: {
    height: 52,
    borderRadius: 26,
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 16,
    ...Platform.select({
      web: { boxShadow: '0px 4px 6px rgba(0, 0, 0, 0.2)' } as any,
      default: {
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.2,
        shadowRadius: 6,
        elevation: 4,
      },
    }),
  },
  ctaText: {
    color: '#FFFFFF',
    fontSize: 17,
    fontWeight: '700',
  },
  ctaSubtext: {
    fontSize: 12,
    textAlign: 'center',
    marginTop: 6,
  },
  secondaryButton: {
    alignItems: 'center',
    paddingVertical: 12,
  },
  secondaryText: {
    fontSize: 13,
    fontWeight: '500',
  },
  footerSection: {
    alignItems: 'center',
    marginTop: 16,
    gap: 10,
  },
  restoreButton: {
    paddingVertical: 6,
    ...Platform.select({
      web: { cursor: 'pointer' } as any,
    }),
  },
  restoreText: {
    fontSize: 13,
    fontWeight: '600',
    textDecorationLine: 'underline',
  },
  legalText: {
    fontSize: 11,
    textAlign: 'center',
    lineHeight: 16,
    paddingHorizontal: 12,
  },
  feedbackBanner: {
    marginHorizontal: 20,
    marginBottom: 12,
    borderRadius: 10,
    padding: 12,
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
  },
  feedbackBannerSuccess: {
    backgroundColor: 'rgba(34, 197, 94, 0.15)',
    borderColor: '#22c55e',
  },
  feedbackBannerError: {
    backgroundColor: 'rgba(239, 68, 68, 0.15)',
    borderColor: '#ef4444',
  },
  feedbackBannerInfo: {
    backgroundColor: 'rgba(59, 130, 246, 0.15)',
    borderColor: '#3b82f6',
  },
  feedbackIcon: {
    fontSize: 20,
    marginRight: 10,
  },
  feedbackTextCol: {
    flex: 1,
  },
  feedbackTitle: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '700',
    marginBottom: 2,
  },
  feedbackMessage: {
    color: '#d1d5db',
    fontSize: 12,
    lineHeight: 16,
  },
  feedbackCloseBtn: {
    padding: 4,
    marginLeft: 8,
    ...Platform.select({
      web: { cursor: 'pointer' } as any,
    }),
  },
  feedbackCloseText: {
    color: '#9ca3af',
    fontSize: 14,
    fontWeight: 'bold',
  },
});
