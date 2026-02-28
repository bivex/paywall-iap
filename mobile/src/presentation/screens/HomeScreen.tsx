import React, {useEffect} from 'react';
import {View, Text, StyleSheet, ScrollView} from 'react-native';
import {useAuthStore} from '../../application/store/authStore';
import {useSubscriptionStore} from '../../application/store/subscriptionStore';
import {navigateToPaywall} from '../navigation/types';

export function HomeScreen() {
  const {user} = useAuthStore();
  const {subscription, checkAccess} = useSubscriptionStore();

  useEffect(() => {
    // Check access on mount
    checkAccess('premium_content');
  }, [checkAccess]);

  const handlePremiumContentAccess = async () => {
    const access = await checkAccess('premium_content');
    if (!access.hasAccess) {
      navigateToPaywall('premium', 'feature_locked');
    }
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Welcome</Text>
        {user && <Text style={styles.subtitle}>User ID: {user.id.slice(0, 8)}...</Text>}
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Subscription Status</Text>
        {subscription ? (
          <View style={styles.statusContainer}>
            <Text style={styles.statusText}>
              Status: <Text style={[styles.statusValue, {color: subscription.status === 'active' ? '#4CAF50' : '#f44336'}]}>
                {subscription.status.toUpperCase()}
              </Text>
            </Text>
            <Text style={styles.statusText}>
              Plan: <Text style={styles.statusValue}>{subscription.planType.toUpperCase()}</Text>
            </Text>
            {subscription.expiresAt && (
              <Text style={styles.statusText}>
                Expires: <Text style={styles.statusValue}>{new Date(subscription.expiresAt).toLocaleDateString()}</Text>
              </Text>
            )}
          </View>
        ) : (
          <Text style={styles.noSubscriptionText}>No active subscription</Text>
        )}
      </View>

      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Content</Text>
        <TouchableOpacity style={styles.contentItem} onPress={handlePremiumContentAccess}>
          <Text style={styles.contentTitle}>Premium Content</Text>
          <Text style={styles.contentDescription}>Exclusive premium content for subscribers</Text>
        </TouchableOpacity>
      </View>
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
  noSubscriptionText: {
    color: '#888',
    fontSize: 14,
  },
  contentItem: {
    backgroundColor: '#1a1a1a',
    borderRadius: 8,
    padding: 15,
  },
  contentTitle: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    marginBottom: 5,
  },
  contentDescription: {
    color: '#888',
    fontSize: 14,
  },
});
