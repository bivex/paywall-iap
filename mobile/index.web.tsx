/**
 * Copyright (c) 2026 Bivex
 * React Native for Web - Application Entrypoint
 */

import React, { useState } from 'react';
import { AppRegistry, StyleSheet, View, Text, TouchableOpacity } from 'react-native';
import App from './App';

function WebContainer() {
  const [deviceFrame, setDeviceFrame] = useState<'iphone' | 'fullscreen'>('iphone');

  return (
    <View style={styles.outerContainer}>
      <View style={styles.topToolbar}>
        <View style={styles.brandRow}>
          <Text style={styles.brandTitle}>Paywall IAP Mobile Web Runner</Text>
          <View style={styles.badge}>
            <Text style={styles.badgeText}>React Native Web</Text>
          </View>
        </View>
        <View style={styles.frameToggleRow}>
          <TouchableOpacity
            style={[styles.toggleBtn, deviceFrame === 'iphone' && styles.toggleBtnActive]}
            onPress={() => setDeviceFrame('iphone')}
          >
            <Text style={[styles.toggleText, deviceFrame === 'iphone' && styles.toggleTextActive]}>
              📱 Phone Frame
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.toggleBtn, deviceFrame === 'fullscreen' && styles.toggleBtnActive]}
            onPress={() => setDeviceFrame('fullscreen')}
          >
            <Text style={[styles.toggleText, deviceFrame === 'fullscreen' && styles.toggleTextActive]}>
              🖥️ Full Viewport
            </Text>
          </TouchableOpacity>
        </View>
      </View>

      <View style={styles.viewportArea}>
        {deviceFrame === 'iphone' ? (
          <View style={styles.phoneFrame}>
            <View style={styles.dynamicIsland} />
            <View style={styles.phoneScreen}>
              <App />
            </View>
            <View style={styles.homeIndicator} />
          </View>
        ) : (
          <View style={styles.fullscreenContent}>
            <App />
          </View>
        )}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  outerContainer: {
    flex: 1,
    height: '100vh' as any,
    width: '100vw' as any,
    backgroundColor: '#090d16',
    flexDirection: 'column',
    overflow: 'hidden',
  },
  topToolbar: {
    height: 54,
    backgroundColor: '#111827',
    borderBottomWidth: 1,
    borderBottomColor: '#1f2937',
    paddingHorizontal: 20,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    zIndex: 100,
  },
  brandRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 12,
  },
  brandTitle: {
    color: '#f9fafb',
    fontSize: 15,
    fontWeight: '700',
    fontFamily: 'system-ui, -apple-system, sans-serif',
  },
  badge: {
    backgroundColor: '#3730a3',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: 6,
  },
  badgeText: {
    color: '#c7d2fe',
    fontSize: 11,
    fontWeight: '600',
  },
  frameToggleRow: {
    flexDirection: 'row',
    backgroundColor: '#1f2937',
    borderRadius: 8,
    padding: 3,
    gap: 4,
  },
  toggleBtn: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
  },
  toggleBtnActive: {
    backgroundColor: '#4f46e5',
  },
  toggleText: {
    color: '#9ca3af',
    fontSize: 12,
    fontWeight: '600',
  },
  toggleTextActive: {
    color: '#ffffff',
  },
  viewportArea: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 20,
    overflow: 'auto' as any,
  },
  phoneFrame: {
    width: 390,
    height: 810,
    maxHeight: '92vh' as any,
    backgroundColor: '#000000',
    borderRadius: 48,
    borderWidth: 8,
    borderColor: '#262626',
    overflow: 'hidden',
    position: 'relative',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 20 },
    shadowOpacity: 0.6,
    shadowRadius: 35,
  },
  dynamicIsland: {
    position: 'absolute',
    top: 10,
    left: '50%' as any,
    transform: [{ translateX: -60 }],
    width: 120,
    height: 28,
    backgroundColor: '#000000',
    borderRadius: 20,
    zIndex: 9999,
  },
  homeIndicator: {
    position: 'absolute',
    bottom: 8,
    left: '50%' as any,
    transform: [{ translateX: -65 }],
    width: 130,
    height: 5,
    backgroundColor: '#ffffff55',
    borderRadius: 3,
    zIndex: 9999,
  },
  phoneScreen: {
    flex: 1,
    backgroundColor: '#000000',
  },
  fullscreenContent: {
    flex: 1,
    width: '100%',
    maxWidth: 600,
    height: '100%',
    backgroundColor: '#000000',
  },
});

AppRegistry.registerComponent('iap-mobile', () => WebContainer);

const rootTag = document.getElementById('root');
if (rootTag) {
  AppRegistry.runApplication('iap-mobile', { rootTag });
}
