/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-02-28 19:38
 * Last Updated: 2026-02-28 19:38
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

import React, {useEffect, useState} from 'react';
import {View, Text, StyleSheet, TouchableOpacity, ActivityIndicator} from 'react-native';
import {useNavigation, useRoute} from '@react-navigation/native';
import {useAuthStore} from '../../application/store/authStore';

interface RegisterParams {
  deviceId: string;
  platform: 'ios' | 'android';
  appVersion: string;
}

export function RegisterScreen() {
  const navigation = useNavigation();
  const route = useRoute();
  const params = (route.params as RegisterParams) || {};

  const {register, isLoading, error} = useAuthStore();
  const [isRegistering, setIsRegistering] = useState(false);

  useEffect(() => {
    // Auto-register on mount with device info
    handleRegister();
  }, []);

  const handleRegister = async () => {
    try {
      setIsRegistering(true);
      await register(
        params.deviceId || 'unknown',
        params.platform || 'ios',
        params.appVersion || '1.0.0',
      );
      // Navigation will happen automatically via auth state change
    } catch (error) {
      console.error('[Register] Failed:', error);
    } finally {
      setIsRegistering(false);
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.content}>
        {(isLoading || isRegistering) ? (
          <>
            <ActivityIndicator size="large" color="#4CAF50" />
            <Text style={styles.statusText}>Setting up your account...</Text>
          </>
        ) : error ? (
          <>
            <Text style={styles.errorTitle}>Setup Failed</Text>
            <Text style={styles.errorMessage}>{error}</Text>
            <TouchableOpacity style={styles.retryButton} onPress={handleRegister}>
              <Text style={styles.retryButtonText}>Retry</Text>
            </TouchableOpacity>
          </>
        ) : (
          <>
            <Text style={styles.successIcon}>✓</Text>
            <Text style={styles.successTitle}>Welcome!</Text>
            <Text style={styles.successMessage}>Your account is ready</Text>
          </>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 30,
  },
  statusText: {
    color: '#fff',
    marginTop: 20,
    fontSize: 16,
  },
  errorTitle: {
    color: '#f44336',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 15,
    textAlign: 'center',
  },
  errorMessage: {
    color: '#fff',
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 30,
  },
  retryButton: {
    backgroundColor: '#4CAF50',
    borderRadius: 12,
    padding: 15,
    paddingHorizontal: 30,
  },
  retryButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  successIcon: {
    fontSize: 60,
    color: '#4CAF50',
    marginBottom: 20,
  },
  successTitle: {
    color: '#fff',
    fontSize: 28,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  successMessage: {
    color: '#888',
    fontSize: 16,
  },
});
