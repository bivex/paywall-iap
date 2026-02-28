import React, {useEffect, useState} from 'react';
import {View, Text, StyleSheet, ScrollView, TouchableOpacity, ActivityIndicator} from 'react-native';
import {useNavigation, RouteProp} from '@react-navigation/native';
import {useIAPStore} from '../../application/store/iapStore';
import {useAuthStore} from '../../application/store/authStore';
import {useSubscriptionStore} from '../../application/store/subscriptionStore';
import {PaywallType, PAYWALL_CONFIGS} from '../../domain/entities/Paywall';
import type {RootStackParamList} from '../navigation/types';

type PaywallScreenRouteProp = RouteProp<RootStackParamList, 'Paywall'>;

interface Props {
  route: PaywallScreenRouteProp;
}

export function PaywallScreen({route}: Props) {
  const {paywallType, trigger} = route.params;
  const navigation = useNavigation();
  const {products, isPurchasing, error: iapError, purchase, finishPurchase} = useIAPStore();
  const {user} = useAuthStore();
  const {subscription, fetchSubscription} = useSubscriptionStore();

  const [selectedProductId, setSelectedProductId] = useState<string | null>(null);

  const config = PAYWALL_CONFIGS[paywallType];

  useEffect(() => {
    // Track paywall view
    console.log('[Paywall] Viewed:', {paywallType, trigger, userId: user?.id});
  }, [paywallType, trigger, user?.id]);

  const handlePurchase = async (productId: string) => {
    try {
      setSelectedProductId(productId);
      const purchase = await purchase(productId);

      // Finish the purchase (verifies with backend)
      await finishPurchase(purchase);

      // Refresh subscription status
      await fetchSubscription();

      // Navigate back or to home
      navigation.goBack();
    } catch (error) {
      console.error('[Paywall] Purchase failed:', error);
    } finally {
      setSelectedProductId(null);
    }
  };

  const handleRestore = async () => {
    try {
      const {restorePurchases} = useIAPStore.getState();
      await restorePurchases();
      await fetchSubscription();
    } catch (error) {
      console.error('[Paywall] Restore failed:', error);
    }
  };

  const handleClose = () => {
    navigation.goBack();
  };

  const getProductForPrice = (priceString: string) => {
    // Match product based on price string
    // This is a simplified approach - real implementation would be more sophisticated
    return products.find(p => p.price.includes(priceString.replace(/[^0-9.]/g, ''))) || products[0];
  };

  return (
    <View style={styles.container}>
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {/* Header */}
        <View style={styles.header}>
          <TouchableOpacity onPress={handleClose} style={styles.closeButton}>
            <Text style={styles.closeButtonText}>✕</Text>
          </TouchableOpacity>
          <Text style={styles.title}>{config.title}</Text>
        </View>

        {/* Features */}
        <View style={styles.featuresContainer}>
          {config.features.map((feature, index) => (
            <View key={index} style={styles.featureItem}>
              <Text style={styles.featureBullet}>✓</Text>
              <Text style={styles.featureText}>{feature}</Text>
            </View>
          ))}
        </View>

        {/* Pricing Options */}
        <View style={styles.pricingContainer}>
          <TouchableOpacity
            style={[styles.pricingOption, styles.monthlyOption]}
            onPress={() => {
              const product = getProductForPrice(config.priceMonthly);
              if (product) handlePurchase(product.productId);
            }}
            disabled={isPurchasing}>
            <Text style={styles.priceLabel}>Monthly</Text>
            <Text style={styles.price}>{config.priceMonthly}</Text>
            {isPurchasing && selectedProductId === getProductForPrice(config.priceMonthly)?.productId ? (
              <ActivityIndicator style={styles.spinner} />
            ) : (
              <Text style={styles.ctaButton}>{config.ctaText}</Text>
            )}
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.pricingOption, styles.annualOption]}
            onPress={() => {
              const product = getProductForPrice(config.priceAnnual);
              if (product) handlePurchase(product.productId);
            }}
            disabled={isPurchasing}>
            <View style={styles.badge}>
              <Text style={styles.badgeText}>SAVE 33%</Text>
            </View>
            <Text style={styles.priceLabel}>Annual</Text>
            <Text style={styles.price}>{config.priceAnnual}</Text>
            {isPurchasing && selectedProductId === getProductForPrice(config.priceAnnual)?.productId ? (
              <ActivityIndicator style={styles.spinner} />
            ) : (
              <Text style={styles.ctaButton}>{config.ctaText}</Text>
            )}
          </TouchableOpacity>
        </View>

        {/* Restore Purchases */}
        <TouchableOpacity onPress={handleRestore} style={styles.restoreButton}>
          <Text style={styles.restoreButtonText}>Restore Purchases</Text>
        </TouchableOpacity>

        {/* Error Message */}
        {iapError && (
          <Text style={styles.errorText}>{iapError}</Text>
        )}

        {/* Terms */}
        <Text style={styles.termsText}>
          Subscription auto-renews unless canceled. Cancel anytime in Settings.
        </Text>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  scrollContent: {
    padding: 20,
    paddingBottom: 40,
  },
  header: {
    alignItems: 'center',
    marginBottom: 30,
  },
  closeButton: {
    position: 'absolute',
    right: 0,
    top: 0,
    padding: 10,
  },
  closeButtonText: {
    color: '#fff',
    fontSize: 24,
  },
  title: {
    color: '#fff',
    fontSize: 28,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  featuresContainer: {
    marginBottom: 30,
  },
  featureItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 15,
  },
  featureBullet: {
    color: '#4CAF50',
    fontSize: 20,
    marginRight: 10,
  },
  featureText: {
    color: '#fff',
    fontSize: 16,
  },
  pricingContainer: {
    marginBottom: 20,
  },
  pricingOption: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 20,
    marginBottom: 15,
    alignItems: 'center',
  },
  monthlyOption: {
    borderWidth: 1,
    borderColor: '#333',
  },
  annualOption: {
    borderWidth: 2,
    borderColor: '#FFD700',
  },
  badge: {
    position: 'absolute',
    top: -10,
    right: 10,
    backgroundColor: '#FFD700',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  badgeText: {
    color: '#000',
    fontSize: 12,
    fontWeight: 'bold',
  },
  priceLabel: {
    color: '#888',
    fontSize: 14,
    marginBottom: 5,
  },
  price: {
    color: '#fff',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 15,
  },
  ctaButton: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  spinner: {
    marginTop: 10,
  },
  restoreButton: {
    padding: 15,
    alignItems: 'center',
  },
  restoreButtonText: {
    color: '#888',
    fontSize: 14,
  },
  errorText: {
    color: '#f44336',
    textAlign: 'center',
    marginTop: 15,
  },
  termsText: {
    color: '#666',
    fontSize: 12,
    textAlign: 'center',
    marginTop: 20,
  },
});
