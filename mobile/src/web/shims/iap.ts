/**
 * Web shim for react-native-iap
 * Allows full purchase testing, cancellation simulation, and receipt generation in web browser.
 */

export interface Product {
  productId: string;
  price: string;
  currency: string;
  title: string;
  description: string;
  localizedPrice?: string;
}

export interface Purchase {
  productId: string;
  transactionId: string;
  transactionReceipt: string;
  transactionDate: number;
}

export interface SubscriptionPurchase extends Purchase {
  autoRenewingAndroid?: boolean;
  originalTransactionDateIOS?: string;
  originalTransactionIdentifierIOS?: string;
}

export interface SubscriptionAndroid extends Product {
  subscriptionOfferDetails?: any[];
}

const mockProducts: Product[] = [
  {
    productId: 'com.mothsalt.game1.yearly',
    price: '$59.99',
    currency: 'USD',
    title: 'Annual Pro',
    description: 'Billed annually',
  },
  {
    productId: 'com.mothsalt.game1.monthly',
    price: '$9.99',
    currency: 'USD',
    title: 'Monthly Pro',
    description: 'Billed monthly',
  },
  {
    productId: 'com.yourapp.premium.yearly',
    price: '$59.99',
    currency: 'USD',
    title: 'Annual Pro',
    description: 'Billed annually',
  },
  {
    productId: 'com.yourapp.premium.monthly',
    price: '$9.99',
    currency: 'USD',
    title: 'Monthly Pro',
    description: 'Billed monthly',
  },
  {
    productId: 'com.yourapp.premium_plus.yearly',
    price: '$99.99',
    currency: 'USD',
    title: 'Annual Pro Plus',
    description: 'Billed annually',
  },
  {
    productId: 'com.yourapp.premium_plus.monthly',
    price: '$14.99',
    currency: 'USD',
    title: 'Monthly Pro Plus',
    description: 'Billed monthly',
  },
];

const normalizeSkus = (input: any): string[] => {
  if (Array.isArray(input)) return input;
  if (input && Array.isArray(input.skus)) return input.skus;
  return [];
};

const normalizeSku = (input: any): string => {
  if (typeof input === 'string') return input;
  if (input && typeof input.sku === 'string') return input.sku;
  return 'com.mothsalt.game1.yearly';
};

export const initConnection = async (): Promise<boolean> => {
  console.log('[react-native-iap:Web] Connection initialized');
  return true;
};

export const endConnection = async (): Promise<void> => {
  console.log('[react-native-iap:Web] Connection ended');
};

export const getProducts = async (input: any): Promise<Product[]> => {
  const skus = normalizeSkus(input);
  if (skus.length === 0) return mockProducts;
  return mockProducts.filter((p) => skus.includes(p.productId));
};

export const getSubscriptions = async (input: any): Promise<Product[]> => {
  const skus = normalizeSkus(input);
  if (skus.length === 0) return mockProducts;
  return mockProducts.filter((p) => skus.includes(p.productId));
};

export type MockIAPScenario = 'valid_active' | 'expired' | 'pending' | 'invalid' | 'canceled_active_';

export const getMockScenario = (): MockIAPScenario => {
  if (typeof localStorage !== 'undefined') {
    return (localStorage.getItem('__web_mock_scenario') as MockIAPScenario) || 'valid_active';
  }
  return 'valid_active';
};

export const setMockScenario = (scenario: MockIAPScenario): void => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem('__web_mock_scenario', scenario);
    console.log('[react-native-iap:Web] Mock scenario set to:', scenario);
  }
};

export const requestPurchase = async (input: any): Promise<Purchase> => {
  const sku = normalizeSku(input);
  const scenario = getMockScenario();
  const txId = 'web_tx_' + Date.now();
  const token = `${scenario}_tok_${Date.now()}`;
  const receipt = JSON.stringify({
    packageName: 'com.mothsalt.game1',
    productId: sku,
    purchaseToken: token,
    type: 'subscription',
  });
  console.log(`[react-native-iap:Web] Simulating purchase for: ${sku} with scenario: ${scenario}`);
  const purchase: Purchase = {
    productId: sku,
    transactionId: txId,
    transactionReceipt: receipt,
    transactionDate: Date.now(),
  };

  try {
    if (typeof localStorage !== 'undefined') {
      if (scenario === 'valid_active' || scenario === 'canceled_active_') {
        const existing = JSON.parse(localStorage.getItem('__web_active_purchases') || '[]');
        const updated = [purchase, ...existing.filter((p: any) => p.productId !== sku)];
        localStorage.setItem('__web_active_purchases', JSON.stringify(updated));

        const history = JSON.parse(localStorage.getItem('__web_receipt_history') || '[]');
        const updatedHistory = [purchase, ...history.filter((p: any) => p.productId !== sku)];
        localStorage.setItem('__web_receipt_history', JSON.stringify(updatedHistory));
      }
    }
  } catch {}

  return purchase;
};

export const requestSubscription = async (input: any): Promise<Purchase> => {
  return requestPurchase(input);
};

export const finishTransaction = async (arg: any, isConsumable?: boolean): Promise<void> => {
  const txId = arg?.purchase?.transactionId || arg?.transactionId || 'web_tx';
  console.log('[react-native-iap:Web] Transaction finished:', txId);
  const purchase = arg?.purchase || (arg?.transactionId ? arg : null);
  if (purchase && typeof localStorage !== 'undefined') {
    try {
      const existing = JSON.parse(localStorage.getItem('__web_active_purchases') || '[]');
      const updated = [purchase, ...existing.filter((p: any) => p.transactionId !== txId)];
      localStorage.setItem('__web_active_purchases', JSON.stringify(updated));

      const history = JSON.parse(localStorage.getItem('__web_receipt_history') || '[]');
      const updatedHistory = [purchase, ...history.filter((p: any) => p.transactionId !== txId)];
      localStorage.setItem('__web_receipt_history', JSON.stringify(updatedHistory));
    } catch {}
  }
};

export const getAvailablePurchases = async (): Promise<Purchase[]> => {
  if (typeof localStorage === 'undefined') return [];
  const active = localStorage.getItem('__web_active_purchases');
  if (active) {
    try {
      const parsed = JSON.parse(active);
      if (Array.isArray(parsed) && parsed.length > 0) return parsed;
    } catch {}
  }
  const history = localStorage.getItem('__web_receipt_history');
  if (history) {
    try {
      const parsed = JSON.parse(history);
      if (Array.isArray(parsed) && parsed.length > 0) return parsed;
    } catch {}
  }
  return [];
};

export const getSubscriptionPurchases = async (): Promise<SubscriptionPurchase[]> => {
  return [];
};

export const flushFailedPurchasesIOS = async (): Promise<void> => {};

export const purchaseUpdatedListener = (listener: (purchase: Purchase) => void) => {
  return {
    remove: () => {},
  };
};

export const purchaseErrorListener = (listener: (error: any) => void) => {
  return {
    remove: () => {},
  };
};

export default {
  initConnection,
  endConnection,
  getProducts,
  getSubscriptions,
  requestPurchase,
  requestSubscription,
  finishTransaction,
  getAvailablePurchases,
  getSubscriptionPurchases,
  flushFailedPurchasesIOS,
  purchaseUpdatedListener,
  purchaseErrorListener,
  getMockScenario,
  setMockScenario,
};
