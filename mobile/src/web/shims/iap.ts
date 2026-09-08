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

export const requestPurchase = async (input: any): Promise<Purchase> => {
  const sku = normalizeSku(input);
  const txId = 'web_tx_' + Date.now();
  const token = 'valid_active_' + Date.now();
  const receipt = JSON.stringify({
    packageName: 'com.mothsalt.game1',
    productId: sku,
    purchaseToken: token,
    type: 'subscription',
  });
  console.log('[react-native-iap:Web] Simulating purchase for:', sku);
  return {
    productId: sku,
    transactionId: txId,
    transactionReceipt: receipt,
    transactionDate: Date.now(),
  };
};

export const requestSubscription = async (input: any): Promise<Purchase> => {
  return requestPurchase(input);
};

export const finishTransaction = async (arg: any, isConsumable?: boolean): Promise<void> => {
  const txId = arg?.purchase?.transactionId || arg?.transactionId || 'web_tx';
  console.log('[react-native-iap:Web] Transaction finished:', txId);
};

export const getAvailablePurchases = async (): Promise<Purchase[]> => {
  const saved = localStorage.getItem('__web_active_purchases');
  if (!saved) return [];
  try {
    return JSON.parse(saved);
  } catch {
    return [];
  }
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
};
