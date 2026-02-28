import React from 'react';
import {View, Text, StyleSheet, ScrollView, TouchableOpacity} from 'react-native';
import {useSubscriptionStore} from '../../application/store/subscriptionStore';
import {navigateToPaywall} from '../navigation/types';

export function SubscriptionScreen() {
  const {subscription, isLoading, error} = useSubscriptionStore();

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
            <Text style={styles.planType}>{subscription.planType.toUpperCase()}</Text>
            <Text style={styles.statusText}>
              Status: <Text style={{color: subscription.status === 'active' ? '#4CAF50' : '#f44336'}}>
                {subscription.status.toUpperCase()}
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
              <Text style={styles.platformText}>{subscription.platform.toUpperCase()}</Text>
              {subscription.source && (
                <Text style={styles.sourceText}>via {subscription.source.toUpperCase()}</Text>
              )}
            </View>
          )}

          <TouchableOpacity
            style={styles.button}
            onPress={() => navigateToPaywall('premium_plus', 'upgrade_prompt')}>
            <Text style={styles.buttonText}>Upgrade Plan</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.button, styles.secondaryButton]}
            onPress={() => {/* TODO: Cancel subscription */}}>
            <Text style={styles.buttonText}>Cancel Subscription</Text>
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
