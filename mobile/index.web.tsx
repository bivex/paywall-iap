/**
 * Copyright (c) 2026 Bivex
 * React Native for Web - Application Entrypoint
 */

import React, { useState } from 'react';
import { AppRegistry, StyleSheet, View, Text, TouchableOpacity, Platform } from 'react-native';
import App from './App';
import { getMockScenario, setMockScenario, MockIAPScenario } from './src/web/shims/iap';

function WebContainer() {
  const [deviceFrame, setDeviceFrame] = useState<'iphone' | 'fullscreen'>('iphone');
  const [scenario, setScenario] = useState<MockIAPScenario>(getMockScenario());
  const [webhookStatus, setWebhookStatus] = useState<string | null>(null);

  const handleSetScenario = (s: MockIAPScenario) => {
    setMockScenario(s);
    setScenario(s);
  };

  const sendGoogleWebhook = async (notificationType: number, label: string) => {
    setWebhookStatus(`Sending ${label}...`);
    try {
      const activePurchases = JSON.parse(localStorage.getItem('__web_active_purchases') || '[]');
      const historyPurchases = JSON.parse(localStorage.getItem('__web_receipt_history') || '[]');
      const last = activePurchases[0] || historyPurchases[0];
      let token = 'valid_active_tok_web';
      let productId = 'com.mothsalt.game1.monthly';
      if (last?.transactionReceipt) {
        try {
          const parsed = JSON.parse(last.transactionReceipt);
          if (parsed.purchaseToken) token = parsed.purchaseToken;
          if (parsed.productId) productId = parsed.productId;
        } catch {}
      }

      const payload = {
        version: '1.0',
        packageName: 'com.mothsalt.game1',
        eventTimeMillis: String(Date.now()),
        subscriptionNotification: {
          version: '1.0',
          notificationType,
          purchaseToken: token,
          subscriptionId: productId,
        },
      };

      const base64Data = btoa(JSON.stringify(payload));
      const res = await fetch('/webhook/google', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: {
            data: base64Data,
            messageId: 'web-msg-' + Date.now(),
          },
          subscription: 'projects/test/subscriptions/iap',
        }),
      });

      if (res.ok) {
        setWebhookStatus(`✅ ${label} sent!`);
        try {
          const { useSubscriptionStore } = await import('./src/application/store/subscriptionStore');
          await useSubscriptionStore.getState().fetchSubscription();
          await useSubscriptionStore.getState().checkAccess('premium_content');
        } catch {}
      } else {
        setWebhookStatus(`❌ ${label} failed: HTTP ${res.status}`);
      }
    } catch (err: any) {
      setWebhookStatus(`❌ Error: ${err.message}`);
    } finally {
      setTimeout(() => setWebhookStatus(null), 3000);
    }
  };

  const refreshState = async () => {
    try {
      const { useSubscriptionStore } = await import('./src/application/store/subscriptionStore');
      await useSubscriptionStore.getState().fetchSubscription();
      await useSubscriptionStore.getState().checkAccess('premium_content');
      setWebhookStatus('🔄 State refreshed');
      setTimeout(() => setWebhookStatus(null), 2000);
    } catch {}
  };

  return (
    <View style={styles.outerContainer}>
      {/* Top Main Toolbar */}
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

      {/* Sub-Toolbar: IAP Mock Scenario Selector & Webhook Testing */}
      <View style={styles.subToolbar}>
        <View style={styles.toolbarSection}>
          <Text style={styles.toolbarLabel}>🧪 Mock IAP Scenario:</Text>
          <View style={styles.btnGroup}>
            <TouchableOpacity
              style={[styles.scenarioBtn, scenario === 'valid_active' && styles.scenarioBtnActive]}
              onPress={() => handleSetScenario('valid_active')}
            >
              <Text style={[styles.scenarioBtnText, scenario === 'valid_active' && styles.scenarioBtnTextActive]}>
                ✅ Success (Active)
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.scenarioBtn, scenario === 'expired' && styles.scenarioBtnActive]}
              onPress={() => handleSetScenario('expired')}
            >
              <Text style={[styles.scenarioBtnText, scenario === 'expired' && styles.scenarioBtnTextActive]}>
                ⏱️ Expired (422)
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.scenarioBtn, scenario === 'invalid' && styles.scenarioBtnActive]}
              onPress={() => handleSetScenario('invalid')}
            >
              <Text style={[styles.scenarioBtnText, scenario === 'invalid' && styles.scenarioBtnTextActive]}>
                ❌ Decline (410)
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.scenarioBtn, scenario === 'canceled_active_' && styles.scenarioBtnActive]}
              onPress={() => handleSetScenario('canceled_active_')}
            >
              <Text style={[styles.scenarioBtnText, scenario === 'canceled_active_' && styles.scenarioBtnTextActive]}>
                🔄 Cancelled + Paid
              </Text>
            </TouchableOpacity>
          </View>
        </View>

        <View style={styles.toolbarSection}>
          <Text style={styles.toolbarLabel}>⚡ RTDN Webhooks:</Text>
          <View style={styles.btnGroup}>
            <TouchableOpacity
              style={styles.webhookBtn}
              onPress={() => sendGoogleWebhook(2, 'Renew')}
            >
              <Text style={styles.webhookBtnText}>🔔 Auto-Renew</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.webhookBtn}
              onPress={() => sendGoogleWebhook(6, 'Grace Period')}
            >
              <Text style={styles.webhookBtnText}>⚠️ Grace</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.webhookBtn}
              onPress={() => sendGoogleWebhook(13, 'Expire')}
            >
              <Text style={styles.webhookBtnText}>🚫 Expire</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.webhookBtn, styles.refreshBtn]}
              onPress={refreshState}
            >
              <Text style={styles.refreshBtnText}>🔄 Refresh</Text>
            </TouchableOpacity>
          </View>
        </View>

        {webhookStatus && (
          <View style={styles.statusPill}>
            <Text style={styles.statusPillText}>{webhookStatus}</Text>
          </View>
        )}
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
    height: 50,
    backgroundColor: '#111827',
    borderBottomWidth: 1,
    borderBottomColor: '#1f2937',
    paddingHorizontal: 20,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    zIndex: 100,
  },
  subToolbar: {
    minHeight: 44,
    backgroundColor: '#0f172a',
    borderBottomWidth: 1,
    borderBottomColor: '#1e293b',
    paddingHorizontal: 16,
    paddingVertical: 6,
    flexDirection: 'row',
    alignItems: 'center',
    flexWrap: 'wrap',
    gap: 16,
    zIndex: 99,
  },
  toolbarSection: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  toolbarLabel: {
    color: '#94a3b8',
    fontSize: 12,
    fontWeight: '600',
  },
  btnGroup: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  scenarioBtn: {
    paddingHorizontal: 9,
    paddingVertical: 4,
    borderRadius: 6,
    backgroundColor: '#1e293b',
    borderWidth: 1,
    borderColor: '#334155',
  },
  scenarioBtnActive: {
    backgroundColor: '#3b82f6',
    borderColor: '#60a5fa',
  },
  scenarioBtnText: {
    color: '#cbd5e1',
    fontSize: 11,
    fontWeight: '600',
  },
  scenarioBtnTextActive: {
    color: '#ffffff',
  },
  webhookBtn: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 6,
    backgroundColor: '#1e293b',
    borderWidth: 1,
    borderColor: '#475569',
  },
  webhookBtnText: {
    color: '#cbd5e1',
    fontSize: 11,
    fontWeight: '500',
  },
  refreshBtn: {
    borderColor: '#10b981',
    backgroundColor: '#064e3b',
  },
  refreshBtnText: {
    color: '#34d399',
    fontSize: 11,
    fontWeight: '600',
  },
  statusPill: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    backgroundColor: '#1e293b',
    borderRadius: 6,
  },
  statusPillText: {
    color: '#e2e8f0',
    fontSize: 11,
    fontWeight: '600',
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
    ...Platform.select({
      web: { boxShadow: '0 20px 35px rgba(0, 0, 0, 0.6)' } as any,
    }),
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
