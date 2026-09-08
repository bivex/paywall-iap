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
];

export const initConnection = async (): Promise<boolean> => {
  console.log('[react-native-iap:Web] Connection initialized');
  return true;
};

export const endConnection = async (): Promise<void> => {
  console.log('[react-native-iap:Web] Connection ended');
};

export const getProducts = async ({ skus }: { skus: string[] }): Promise<Product[]> => {
  return mockProducts.filter((p) => skus.includes(p.productId));
};

export const getSubscriptions = async ({ skus }: { skus: string[] }): Promise<Product[]> => {
  return mockProducts.filter((p) => skus.includes(p.productId));
};

export const requestPurchase = async ({ sku }: { sku: string }): Promise<Purchase> => {
  console.log('[react-native-iap:Web] Simulating purchase for:', sku);
  return {
    productId: sku,
    transactionId: 'web_tx_' + Date.now(),
    transactionReceipt: 'web_simulated_receipt_' + btoa(JSON.stringify({ sku, time: Date.now() })),
    transactionDate: Date.now(),
  };
};

export const requestSubscription = async ({ sku }: { sku: string }): Promise<Purchase> => {
  return requestPurchase({ sku });
};

export const finishTransaction = async ({ purchase }: { purchase: Purchase }): Promise<void> => {
  console.log('[react-native-iap:Web] Transaction finished:', purchase.transactionId);
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
