import React, {useState} from 'react';
import {View, Text, StyleSheet, ScrollView, TouchableOpacity, Alert, Platform, ActivityIndicator} from 'react-native';
import {useSubscriptionStore} from '../../application/store/subscriptionStore';
import {navigateToPaywall} from '../navigation/types';

export function SubscriptionScreen() {
  const {subscription, isLoading, error, cancelSubscription} = useSubscriptionStore();
  const [isCancelling, setIsCancelling] = useState(false);
  const [isRestoring, setIsRestoring] = useState(false);

  const handleRestore = async () => {
    setIsRestoring(true);
    try {
      const { PaywallSDK } = await import('../../sdk');
      const res = await PaywallSDK.restorePurchases();
      if (res.success && (res.restoredCount ?? 0) > 0) {
        const msg = 'Your subscription was successfully restored!';
        if (Platform.OS === 'web' && typeof window !== 'undefined') {
          window.alert('Restored: ' + msg);
        } else {
          Alert.alert('Restored', msg);
        }
      } else if (res.success) {
        const msg = 'No previous active purchases were found to restore.';
        if (Platform.OS === 'web' && typeof window !== 'undefined') {
          window.alert('No Purchases: ' + msg);
        } else {
          Alert.alert('No Purchases', msg);
        }
      } else {
        const msg = res.error || 'Could not restore purchases.';
        if (Platform.OS === 'web' && typeof window !== 'undefined') {
          window.alert('Restore Failed: ' + msg);
        } else {
          Alert.alert('Restore Failed', msg);
        }
      }
    } catch (err: any) {
      const msg = err?.message || 'Restore failed';
      if (Platform.OS === 'web' && typeof window !== 'undefined') {
        window.alert('Error: ' + msg);
      } else {
        Alert.alert('Error', msg);
      }
    } finally {
      setIsRestoring(false);
    }
  };

  const handleCancel = () => {
    const performCancel = async () => {
      setIsCancelling(true);
      try {
        await cancelSubscription();
        if (Platform.OS !== 'web') {
          Alert.alert('Subscription Cancelled', 'Your subscription has been successfully cancelled.');
        }
      } catch (err: any) {
        const msg = err?.message || 'Failed to cancel subscription';
        if (Platform.OS === 'web' && typeof window !== 'undefined') {
          window.alert(msg);
        } else {
          Alert.alert('Error', msg);
        }
      } finally {
        setIsCancelling(false);
      }
    };

    if (Platform.OS === 'web' && typeof window !== 'undefined' && window.confirm) {
      if (window.confirm('Are you sure you want to cancel your subscription? You will lose access to premium features.')) {
        performCancel();
      }
    } else {
      Alert.alert(
        'Cancel Subscription',
        'Are you sure you want to cancel your subscription? You will lose access to premium features.',
        [
          {text: 'Keep Subscription', style: 'cancel'},
          {text: 'Cancel Subscription', style: 'destructive', onPress: performCancel},
        ],
      );
    }
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Subscription</Text>
      </View>

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

          <TouchableOpacity
            style={[styles.button, styles.cancelButton]}
            onPress={handleCancel}
            disabled={isCancelling}>
            {isCancelling ? (
              <ActivityIndicator size="small" color="#fff" />
            ) : (
              <Text style={styles.buttonText}>Cancel Subscription</Text>
            )}
          </TouchableOpacity>
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
});
