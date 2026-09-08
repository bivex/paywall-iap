import React, {useEffect, useState} from 'react';
import {View, Text, StyleSheet, ScrollView, TouchableOpacity, Platform, ActivityIndicator, Modal} from 'react-native';
import {useAuthStore} from '../../application/store/authStore';
import {useSubscriptionStore} from '../../application/store/subscriptionStore';
import {navigateToPaywall} from '../navigation/types';
import {PaywallModal} from '../../sdk';

export function HomeScreen() {
  const {user} = useAuthStore();
  const {subscription, checkAccess, fetchSubscription, cancelSubscription} = useSubscriptionStore();
  const [modalVisible, setModalVisible] = useState(false);
  const [vipModalVisible, setVipModalVisible] = useState(false);
  const [isCancelling, setIsCancelling] = useState(false);
  const [confirmingCancel, setConfirmingCancel] = useState(false);
  const [hasContentAccess, setHasContentAccess] = useState<boolean | null>(null);
  const [feedback, setFeedback] = useState<{
    type: 'success' | 'error' | 'info';
    title: string;
    message: string;
  } | null>(null);

  useEffect(() => {
    // Check access on mount
    checkAccess('premium_content').then((res) => {
      setHasContentAccess(res.hasAccess);
    });
  }, [checkAccess, subscription]);

  const handlePremiumContentAccess = async () => {
    const access = await checkAccess('premium_content');
    setHasContentAccess(access.hasAccess);
    if (!access.hasAccess) {
      navigateToPaywall('premium', 'feature_locked');
    } else {
      setVipModalVisible(true);
    }
  };

  const performCancelSubscription = async () => {
    setIsCancelling(true);
    setConfirmingCancel(false);
    setFeedback(null);
    try {
      await cancelSubscription();
      const access = await checkAccess('premium_content');
      setHasContentAccess(access.hasAccess);
      setFeedback({
        type: 'info',
        title: 'Subscription Cancelled',
        message: 'Renewal cancelled. You retain access until the end of your billing cycle.',
      });
    } catch (err: any) {
      setFeedback({
        type: 'error',
        title: 'Cancellation Failed',
        message: err?.message || 'Failed to cancel subscription',
      });
    } finally {
      setIsCancelling(false);
    }
  };

  return (
    <>
      <ScrollView style={styles.container}>
        <View style={styles.header}>
          <Text style={styles.title}>Welcome</Text>
          {user && <Text style={styles.subtitle}>User ID: {user.id.slice(0, 8)}...</Text>}
        </View>

        {/* Inline Feedback Banner */}
        {feedback && (
          <View
            style={[
              styles.feedbackBanner,
              feedback.type === 'success' && styles.feedbackSuccess,
              feedback.type === 'error' && styles.feedbackError,
              feedback.type === 'info' && styles.feedbackInfo,
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
              style={styles.feedbackClose}
              onPress={() => setFeedback(null)}
              hitSlop={{ top: 8, bottom: 8, left: 8, right: 8 }}
            >
              <Text style={styles.feedbackCloseText}>✕</Text>
            </TouchableOpacity>
          </View>
        )}

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Subscription Status</Text>
          {subscription ? (
            <View style={styles.statusContainer}>
              <Text style={styles.statusText}>
                Status: <Text style={[styles.statusValue, {color: (subscription.status || '').toLowerCase() === 'active' ? '#4CAF50' : '#f44336'}]}>
                  {String(subscription.status || 'ACTIVE').toUpperCase()}
                </Text>
              </Text>
              <Text style={styles.statusText}>
                Plan: <Text style={styles.statusValue}>{String(subscription.planType || (subscription as any).plan_type || 'PRO').toUpperCase()}</Text>
              </Text>
              {subscription.expiresAt && (
                <Text style={styles.statusText}>
                  Expires: <Text style={styles.statusValue}>{new Date(subscription.expiresAt).toLocaleDateString()}</Text>
                </Text>
              )}

              {confirmingCancel ? (
                <View style={styles.inlineConfirmBox}>
                  <Text style={styles.inlineConfirmTitle}>Stop Auto-Renewal?</Text>
                  <Text style={styles.inlineConfirmDesc}>
                    You keep access until your expiration date, but subscription will not renew.
                  </Text>
                  <View style={styles.inlineConfirmActions}>
                    <TouchableOpacity
                      style={[styles.inlineConfirmBtn, styles.inlineKeepBtn]}
                      onPress={() => setConfirmingCancel(false)}
                    >
                      <Text style={styles.inlineKeepText}>Keep</Text>
                    </TouchableOpacity>
                    <TouchableOpacity
                      style={[styles.inlineConfirmBtn, styles.inlineDangerBtn]}
                      onPress={performCancelSubscription}
                      disabled={isCancelling}
                    >
                      {isCancelling ? (
                        <ActivityIndicator size="small" color="#fff" />
                      ) : (
                        <Text style={styles.inlineDangerText}>Confirm Cancel</Text>
                      )}
                    </TouchableOpacity>
                  </View>
                </View>
              ) : (
                <TouchableOpacity
                  style={styles.cancelSubButton}
                  onPress={() => setConfirmingCancel(true)}
                  disabled={isCancelling}
                >
                  <Text style={styles.cancelSubText}>Cancel Subscription</Text>
                </TouchableOpacity>
              )}
            </View>
          ) : (
            <View style={styles.statusContainer}>
              <Text style={styles.noSubscriptionText}>No active subscription</Text>
              <TouchableOpacity
                style={styles.proBannerButton}
                onPress={() => setModalVisible(true)}
              >
                <Text style={styles.proBannerText}>✨ Upgrade with Paywall Modal</Text>
              </TouchableOpacity>
            </View>
          )}
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Content</Text>
          <TouchableOpacity
            style={[
              styles.contentItem,
              hasContentAccess && styles.contentItemUnlocked,
            ]}
            onPress={handlePremiumContentAccess}
          >
            <View style={styles.contentHeaderRow}>
              <Text style={styles.contentTitle}>
                {hasContentAccess ? '💎 Premium Content (Unlocked)' : '🔒 Premium Content (Full Screen Paywall)'}
              </Text>
              <View style={[styles.accessBadge, hasContentAccess ? styles.accessBadgeUnlocked : styles.accessBadgeLocked]}>
                <Text style={styles.accessBadgeText}>{hasContentAccess ? 'UNLOCKED' : 'LOCKED'}</Text>
              </View>
            </View>
            <Text style={styles.contentDescription}>
              {hasContentAccess
                ? 'VIP Member Lounge unlocked! Tap to explore your exclusive subscriber features.'
                : 'Exclusive premium features for subscribers. Tap to unlock via paywall.'}
            </Text>
          </TouchableOpacity>
        </View>
      </ScrollView>

      {/* Embedded SDK Paywall Modal */}
      <PaywallModal
        visible={modalVisible}
        trigger="home_banner_click"
        onClose={() => setModalVisible(false)}
        onPurchaseSuccess={async () => {
          setModalVisible(false);
          await fetchSubscription();
          const access = await checkAccess('premium_content');
          setHasContentAccess(access.hasAccess);
        }}
      />

      {/* VIP Premium Lounge Modal */}
      <Modal
        visible={vipModalVisible}
        animationType="slide"
        transparent={true}
        onRequestClose={() => setVipModalVisible(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.vipCard}>
            <View style={styles.vipHeader}>
              <View style={styles.vipTitleRow}>
                <Text style={styles.vipBadge}>VIP ACCESS</Text>
                <Text style={styles.vipPlanName}>{String(subscription?.planType || 'PRO').toUpperCase()}</Text>
              </View>
              <TouchableOpacity
                style={styles.vipCloseBtn}
                onPress={() => setVipModalVisible(false)}
                hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
              >
                <Text style={styles.vipCloseText}>✕</Text>
              </TouchableOpacity>
            </View>

            <Text style={styles.vipTitle}>💎 Premium Content Lounge</Text>
            <Text style={styles.vipSubtitle}>
              Your subscription is active and verified by the billing engine. All features unlocked!
            </Text>

            <View style={styles.vipFeatureList}>
              <View style={styles.vipFeatureItem}>
                <Text style={styles.vipFeatureIcon}>⚡</Text>
                <View style={styles.vipFeatureInfo}>
                  <Text style={styles.vipFeatureName}>Real-time Market Telemetry</Text>
                  <Text style={styles.vipFeatureDesc}>Live analytics pipeline streaming with 0 latency.</Text>
                </View>
                <Text style={styles.vipActivePill}>ACTIVE</Text>
              </View>

              <View style={styles.vipFeatureItem}>
                <Text style={styles.vipFeatureIcon}>☁️</Text>
                <View style={styles.vipFeatureInfo}>
                  <Text style={styles.vipFeatureName}>Unlimited Cloud Sync</Text>
                  <Text style={styles.vipFeatureDesc}>Cross-platform receipts sync & encrypted backup.</Text>
                </View>
                <Text style={styles.vipActivePill}>ACTIVE</Text>
              </View>

              <View style={styles.vipFeatureItem}>
                <Text style={styles.vipFeatureIcon}>🤖</Text>
                <View style={styles.vipFeatureInfo}>
                  <Text style={styles.vipFeatureName}>AI Financial Advisor</Text>
                  <Text style={styles.vipFeatureDesc}>Advanced portfolio forecasting model enabled.</Text>
                </View>
                <Text style={styles.vipActivePill}>ACTIVE</Text>
              </View>
            </View>

            <TouchableOpacity
              style={styles.vipDoneBtn}
              onPress={() => setVipModalVisible(false)}
            >
              <Text style={styles.vipDoneBtnText}>Done</Text>
            </TouchableOpacity>
          </View>
        </View>
      </Modal>
    </>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  header: {
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#333',
  },
  title: {
    color: '#fff',
    fontSize: 28,
    fontWeight: 'bold',
    marginBottom: 5,
  },
  subtitle: {
    color: '#888',
    fontSize: 14,
  },
  section: {
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#333',
  },
  sectionTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 15,
  },
  statusContainer: {
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    padding: 15,
  },
  statusText: {
    color: '#888',
    fontSize: 14,
    marginBottom: 8,
  },
  statusValue: {
    color: '#fff',
    fontWeight: 'bold',
  },
  cancelSubButton: {
    marginTop: 12,
    paddingVertical: 8,
    paddingHorizontal: 12,
    borderRadius: 6,
    borderWidth: 1,
    borderColor: '#f44336',
    alignSelf: 'flex-start',
  },
  cancelSubText: {
    color: '#f44336',
    fontSize: 13,
    fontWeight: '600',
  },
  noSubscriptionText: {
    color: '#888',
    fontSize: 14,
    marginBottom: 10,
  },
  proBannerButton: {
    backgroundColor: '#6366f1',
    borderRadius: 8,
    paddingVertical: 12,
    paddingHorizontal: 16,
    alignItems: 'center',
    marginTop: 6,
  },
  proBannerText: {
    color: '#ffffff',
    fontSize: 15,
    fontWeight: '700',
  },
  contentItem: {
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    padding: 15,
    borderWidth: 1,
    borderColor: '#333',
  },
  contentItemUnlocked: {
    borderColor: '#6366f1',
    backgroundColor: '#15172b',
  },
  contentHeaderRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 5,
  },
  contentTitle: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    flex: 1,
  },
  contentDescription: {
    color: '#888',
    fontSize: 14,
  },
  accessBadge: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 6,
    marginLeft: 8,
  },
  accessBadgeUnlocked: {
    backgroundColor: 'rgba(34, 197, 94, 0.2)',
    borderColor: '#22c55e',
    borderWidth: 1,
  },
  accessBadgeLocked: {
    backgroundColor: 'rgba(156, 163, 175, 0.2)',
    borderColor: '#6b7280',
    borderWidth: 1,
  },
  accessBadgeText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#fff',
  },
  feedbackBanner: {
    margin: 20,
    marginBottom: 0,
    borderRadius: 10,
    padding: 14,
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
  },
  feedbackSuccess: {
    backgroundColor: 'rgba(34, 197, 94, 0.15)',
    borderColor: '#22c55e',
  },
  feedbackError: {
    backgroundColor: 'rgba(239, 68, 68, 0.15)',
    borderColor: '#ef4444',
  },
  feedbackInfo: {
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
  feedbackClose: {
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
  inlineConfirmBox: {
    backgroundColor: '#1f1515',
    borderColor: '#7f1d1d',
    borderWidth: 1,
    borderRadius: 8,
    padding: 12,
    marginTop: 12,
  },
  inlineConfirmTitle: {
    color: '#f87171',
    fontSize: 14,
    fontWeight: '700',
    marginBottom: 4,
  },
  inlineConfirmDesc: {
    color: '#cbd5e1',
    fontSize: 12,
    marginBottom: 10,
    lineHeight: 16,
  },
  inlineConfirmActions: {
    flexDirection: 'row',
    gap: 8,
  },
  inlineConfirmBtn: {
    flex: 1,
    paddingVertical: 8,
    borderRadius: 6,
    alignItems: 'center',
    justifyContent: 'center',
  },
  inlineKeepBtn: {
    backgroundColor: '#334155',
  },
  inlineKeepText: {
    color: '#f1f5f9',
    fontSize: 13,
    fontWeight: '600',
  },
  inlineDangerBtn: {
    backgroundColor: '#dc2626',
  },
  inlineDangerText: {
    color: '#ffffff',
    fontSize: 13,
    fontWeight: '700',
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0,0,0,0.75)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  vipCard: {
    backgroundColor: '#111827',
    borderRadius: 16,
    padding: 24,
    width: '100%',
    maxWidth: 420,
    borderWidth: 1,
    borderColor: '#374151',
    ...Platform.select({
      web: { boxShadow: '0px 10px 25px rgba(0, 0, 0, 0.5)' } as any,
    }),
  },
  vipHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  vipTitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  vipBadge: {
    backgroundColor: '#4f46e5',
    color: '#fff',
    fontSize: 11,
    fontWeight: '800',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 4,
  },
  vipPlanName: {
    color: '#93c5fd',
    fontSize: 12,
    fontWeight: '700',
  },
  vipCloseBtn: {
    padding: 4,
    ...Platform.select({
      web: { cursor: 'pointer' } as any,
    }),
  },
  vipCloseText: {
    color: '#9ca3af',
    fontSize: 16,
    fontWeight: 'bold',
  },
  vipTitle: {
    color: '#ffffff',
    fontSize: 22,
    fontWeight: '800',
    marginBottom: 6,
  },
  vipSubtitle: {
    color: '#9ca3af',
    fontSize: 13,
    lineHeight: 18,
    marginBottom: 20,
  },
  vipFeatureList: {
    gap: 14,
    marginBottom: 24,
  },
  vipFeatureItem: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1f2937',
    borderRadius: 10,
    padding: 12,
    gap: 12,
  },
  vipFeatureIcon: {
    fontSize: 24,
  },
  vipFeatureInfo: {
    flex: 1,
  },
  vipFeatureName: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '700',
    marginBottom: 2,
  },
  vipFeatureDesc: {
    color: '#9ca3af',
    fontSize: 11,
  },
  vipActivePill: {
    color: '#10b981',
    backgroundColor: 'rgba(16, 185, 129, 0.15)',
    fontSize: 10,
    fontWeight: '800',
    paddingHorizontal: 6,
    paddingVertical: 3,
    borderRadius: 4,
  },
  vipDoneBtn: {
    backgroundColor: '#6366f1',
    borderRadius: 10,
    paddingVertical: 12,
    alignItems: 'center',
    ...Platform.select({
      web: { cursor: 'pointer' } as any,
    }),
  },
  vipDoneBtnText: {
    color: '#ffffff',
    fontSize: 15,
    fontWeight: '700',
  },
});
