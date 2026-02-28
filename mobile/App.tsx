/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-02-28 19:37
 * Last Updated: 2026-02-28 19:37
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

import React, {useEffect, useState} from 'react';
import {StatusBar, SafeAreaView, StyleSheet, View, Text, ActivityIndicator} from 'react-native';
import {Navigation} from './src/presentation/navigation/Navigation';
import {initializeServices, getIAPService, cleanupServices} from './src/application/services/Services';
import {useAuthStore} from './src/application/store/authStore';
import {useIAPStore} from './src/application/store/iapStore';

function App() {
  const [isInitializing, setIsInitializing] = useState(true);
  const [initError, setInitError] = useState<string | null>(null);
  const {loadStoredTokens} = useAuthStore();
  const {setService, initialize: initIAP} = useIAPStore();

  useEffect(() => {
    let mounted = true;

    const init = async () => {
      try {
        // Initialize all services
        await initializeServices();

        // Load stored auth tokens
        await loadStoredTokens();

        // Initialize IAP
        const iapService = getIAPService();
        setService(iapService);
        await initIAP();

        if (mounted) {
          setIsInitializing(false);
        }
      } catch (error) {
        console.error('[App] Initialization failed:', error);
        if (mounted) {
          setInitError(error instanceof Error ? error.message : 'Failed to initialize app');
        }
      }
    };

    init();

    return () => {
      mounted = false;
      // Cleanup on unmount
      cleanupServices();
    };
  }, [loadStoredTokens, setService, initIAP]);

  if (isInitializing) {
    return (
      <SafeAreaView style={styles.container}>
        <StatusBar barStyle="light-content" backgroundColor="#000" />
        <View style={styles.centered}>
          <ActivityIndicator size="large" color="#4CAF50" />
          <Text style={styles.loadingText}>Initializing...</Text>
        </View>
      </SafeAreaView>
    );
  }

  if (initError) {
    return (
      <SafeAreaView style={styles.container}>
        <StatusBar barStyle="light-content" backgroundColor="#000" />
        <View style={styles.centered}>
          <Text style={styles.errorTitle}>Initialization Failed</Text>
          <Text style={styles.errorMessage}>{initError}</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <StatusBar barStyle="light-content" backgroundColor="#000" />
      <Navigation />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  loadingText: {
    color: '#fff',
    marginTop: 15,
    fontSize: 16,
  },
  errorTitle: {
    color: '#f44336',
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 10,
  },
  errorMessage: {
    color: '#fff',
    fontSize: 14,
    textAlign: 'center',
  },
});

export default App;
