/**
 * Copyright (c) 2026 Bivex
 * Paywall Mobile SDK - Modal Presentation Component
 */

import React from 'react';
import { Modal, StyleSheet, View } from 'react-native';
import { PaywallView, PaywallViewProps } from './PaywallView';
import { PurchaseResult } from './types';

export interface PaywallModalProps extends PaywallViewProps {
  visible: boolean;
  onClose: () => void;
  animationType?: 'slide' | 'fade' | 'none';
}

export const PaywallModal: React.FC<PaywallModalProps> = ({
  visible,
  onClose,
  animationType = 'slide',
  onPurchaseSuccess,
  ...viewProps
}) => {
  return (
    <Modal
      visible={visible}
      animationType={animationType}
      presentationStyle="fullScreen"
      onRequestClose={onClose}
    >
      <View style={styles.modalContainer}>
        <PaywallView
          {...viewProps}
          onDismiss={onClose}
          onPurchaseSuccess={(result: PurchaseResult) => {
            onPurchaseSuccess?.(result);
            onClose();
          }}
        />
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalContainer: {
    flex: 1,
  },
});
