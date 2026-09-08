import React, {useState} from 'react';
import {View, Text, StyleSheet, ScrollView, TouchableOpacity, Alert, Platform, ActivityIndicator} from 'react-native';
import {useSubscriptionStore} from '../../application/store/subscriptionStore';
import {navigateToPaywall} from '../navigation/types';

export function SubscriptionScreen() {
  const {subscription, isLoading, error, cancelSubscription, fetchSubscription} = useSubscriptionStore();
  const [isCancelling, setIsCancelling] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);
  const [confirmingCancel, setConfirmingCancel] = useState(false);
  const [feedback, setFeedback] = useState<{
    type: 'success' | 'error' | 'info';
    title: string;
    message: string;
  } | null>(null);

  const handleRestore = async () => {
    setIsRestoring(true);
    setFeedback(null);
    try {
      const { PaywallSDK } = await import('../../sdk');
      const res = await PaywallSDK.restorePurchases();
      if (res.success && (res.restoredCount ?? 0) > 0) {
        setFeedback({
          type: 'success',
          title: 'Purchases Restored',
          message: `Successfully restored ${res.restoredCount} subscription(s)! Your access has been updated.`,
        });
        await fetchSubscription();
      } else if (res.success) {
        setFeedback({
          type: 'info',
          title: 'No Purchases Found',
          message: 'No previous active purchases were found to restore for this account.',
        });
      } else {
        setFeedback({
          type: 'error',
          title: 'Restore Failed',
          message: res.error || 'Could not restore purchases.',
        });
      }
    } catch (err: any) {
      setFeedback({
        type: 'error',
        title: 'Restore Error',
        message: err?.message || 'Restore failed',
      });
    } finally {
      setIsRestoring(false);
    }
  };

  const performCancel = async () => {
    setIsCancelling(true);
    setConfirmingCancel(false);
    setFeedback(null);
    try {
      await cancelSubscription();
      setFeedback({
        type: 'info',
        title: 'Subscription Cancelled',
        message: 'Auto-renewal has been stopped. You will retain access until the expiration date.',
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
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Subscription</Text>
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

      {isLoading ? (
        <View style={styles.centered}>
          <Text style={styles.loadingText}>Loading...</Text>
        </View>
      ) : subscription ? (
        <View style={styles.content}>
          <View style={styles.card}>
            <Text style={styles.cardTitle}>Current Plan</Text>
            <Text style={styles.planType}>{String(subscription.planType || (subscription as any).plan_type || 'PRO').toUpperCase()}</Text>
            <Text style={styles.statusText}>
              Status: <Text style={{color: (subscription.status || '').toLowerCase() === 'active' ? '#4CAF50' : '#f44336'}}>
                {String(subscription.status || 'ACTIVE').toUpperCase()}
              </Text>
            </Text>
            {subscription.expiresAt && (
              <Text style={styles.expiryText}>
                Expires: {new Date(subscription.expiresAt).toLocaleDateString()}
              </Text>
            )}
          </View>

          {subscription.platform && (
            <View style={styles.card}>
              <Text style={styles.cardTitle}>Platform</Text>
              <Text style={styles.platformText}>{String(subscription.platform).toUpperCase()}</Text>
              {subscription.source && (
                <Text style={styles.sourceText}>via {String(subscription.source).toUpperCase()}</Text>
              )}
            </View>
          )}

          <TouchableOpacity
            style={styles.button}
            onPress={() => navigateToPaywall('premium_plus', 'upgrade_prompt')}>
            <Text style={styles.buttonText}>Upgrade Plan</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.button, styles.restoreButton]}
            onPress={handleRestore}
            disabled={isRestoring}>
            {isRestoring ? (
              <ActivityIndicator size="small" color="#fff" />
            ) : (
              <Text style={styles.restoreButtonText}>🔄 Restore Purchases</Text>
            )}
          </TouchableOpacity>

          {confirmingCancel ? (
            <View style={styles.confirmBox}>
              <Text style={styles.confirmTitle}>Cancel Subscription?</Text>
              <Text style={styles.confirmDesc}>
                You will lose access to premium features when your current period expires.
              </Text>
              <View style={styles.confirmBtnRow}>
                <TouchableOpacity
                  style={[styles.confirmBtn, styles.confirmCancelBtn]}
                  onPress={() => setConfirmingCancel(false)}
                >
                  <Text style={styles.confirmCancelText}>Keep Plan</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={[styles.confirmBtn, styles.confirmActionBtn]}
                  onPress={performCancel}
                  disabled={isCancelling}
                >
                  {isCancelling ? (
                    <ActivityIndicator size="small" color="#fff" />
                  ) : (
                    <Text style={styles.confirmActionText}>Yes, Cancel</Text>
                  )}
                </TouchableOpacity>
              </View>
            </View>
          ) : (
            <TouchableOpacity
              style={[styles.button, styles.cancelButton]}
              onPress={() => setConfirmingCancel(true)}
              disabled={isCancelling}>
              <Text style={styles.buttonText}>Cancel Subscription</Text>
            </TouchableOpacity>
          )}
        </View>
      ) : (
        <View style={styles.centered}>
          <Text style={styles.noSubscriptionText}>No active subscription</Text>
          <TouchableOpacity
            style={styles.button}
            onPress={() => navigateToPaywall('premium', 'subscription_expired')}>
            <Text style={styles.buttonText}>Get Premium</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.button, styles.restoreButton]}
            onPress={handleRestore}
            disabled={isRestoring}>
            {isRestoring ? (
              <ActivityIndicator size="small" color="#fff" />
            ) : (
              <Text style={styles.restoreButtonText}>🔄 Restore Purchases</Text>
            )}
          </TouchableOpacity>
        </View>
      )}

      {error && (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
        </View>
      )}
    </ScrollView>
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
  },
  content: {
    padding: 20,
  },
  centered: {
    padding: 20,
    alignItems: 'center',
  },
  loadingText: {
    color: '#888',
  },
  noSubscriptionText: {
    color: '#888',
    fontSize: 16,
    marginBottom: 20,
  },
  card: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 20,
    marginBottom: 15,
  },
  cardTitle: {
    color: '#888',
    fontSize: 14,
    marginBottom: 10,
  },
  planType: {
    color: '#fff',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  statusText: {
    color: '#fff',
    fontSize: 16,
    marginBottom: 5,
  },
  expiryText: {
    color: '#888',
    fontSize: 14,
  },
  platformText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  sourceText: {
    color: '#888',
    fontSize: 14,
  },
  button: {
    backgroundColor: '#4CAF50',
    borderRadius: 8,
    padding: 15,
    alignItems: 'center',
    marginTop: 15,
  },
  secondaryButton: {
    backgroundColor: '#333',
  },
  restoreButton: {
    backgroundColor: '#1e293b',
    borderWidth: 1,
    borderColor: '#3b82f6',
  },
  restoreButtonText: {
    color: '#60a5fa',
    fontSize: 15,
    fontWeight: '600',
  },
  cancelButton: {
    backgroundColor: '#d32f2f',
  },
  buttonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  errorContainer: {
    margin: 20,
    padding: 15,
    backgroundColor: 'rgba(244, 67, 54, 0.1)',
    borderRadius: 8,
  },
  errorText: {
    color: '#f44336',
    textAlign: 'center',
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
  confirmBox: {
    backgroundColor: '#1f1515',
    borderColor: '#7f1d1d',
    borderWidth: 1,
    borderRadius: 10,
    padding: 16,
    marginTop: 16,
  },
  confirmTitle: {
    color: '#f87171',
    fontSize: 15,
    fontWeight: '700',
    marginBottom: 6,
  },
  confirmDesc: {
    color: '#cbd5e1',
    fontSize: 13,
    marginBottom: 14,
    lineHeight: 18,
  },
  confirmBtnRow: {
    flexDirection: 'row',
    gap: 10,
  },
  confirmBtn: {
    flex: 1,
    paddingVertical: 10,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  confirmCancelBtn: {
    backgroundColor: '#334155',
  },
  confirmCancelText: {
    color: '#f1f5f9',
    fontSize: 14,
    fontWeight: '600',
  },
  confirmActionBtn: {
    backgroundColor: '#dc2626',
  },
  confirmActionText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '700',
  },
});
