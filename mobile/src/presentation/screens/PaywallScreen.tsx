import React from 'react';
import { useNavigation, RouteProp } from '@react-navigation/native';
import { PaywallView } from '../../sdk/PaywallView';
import { useSubscriptionStore } from '../../application/store/subscriptionStore';
import type { RootStackParamList } from '../navigation/types';

type PaywallScreenRouteProp = RouteProp<RootStackParamList, 'Paywall'>;

interface Props {
  route: PaywallScreenRouteProp;
}

export function PaywallScreen({ route }: Props) {
  const navigation = useNavigation();
  const trigger = route.params?.trigger || 'default';
  const { fetchSubscription } = useSubscriptionStore();

  const handleDismiss = () => {
    navigation.goBack();
  };

  const handlePurchaseSuccess = async () => {
    try {
      await fetchSubscription();
    } catch {
      // ignore
    }
    navigation.goBack();
  };

  return (
    <PaywallView
      trigger={trigger}
      onDismiss={handleDismiss}
      onPurchaseSuccess={handlePurchaseSuccess}
    />
  );
}
