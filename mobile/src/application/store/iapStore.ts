import {create} from 'zustand';
import {IAPProduct, IAPService} from '../../infrastructure/iap/IAPService';
import {SubscriptionPurchase} from 'react-native-iap';

interface IAPState {
  products: IAPProduct[];
  currentPurchase: SubscriptionPurchase | null;
  iapService: IAPService | null;
  isInitializing: boolean;
  isPurchasing: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  setService: (service: IAPService) => void;
  initialize: () => Promise<void>;
  loadProducts: () => Promise<IAPProduct[]>;
  purchase: (productId: string) => Promise<SubscriptionPurchase>;
  finishPurchase: (purchase: SubscriptionPurchase) => Promise<void>;
  restorePurchases: () => Promise<void>;
  setCurrentPurchase: (purchase: SubscriptionPurchase | null) => void;
  clearError: () => void;
}

export const useIAPStore = create<IAPState>()((set, get) => ({
  products: [],
  currentPurchase: null,
  iapService: null,
  isInitializing: false,
  isPurchasing: false,
  isLoading: false,
  error: null,

  setService: (service: IAPService) => set({iapService: service}),

  initialize: async () => {
    const {iapService} = get();
    if (!iapService) {
      set({error: 'IAP Service not initialized'});
      return;
    }

    set({isInitializing: true, error: null});
    try {
      await iapService.initialize();

      // Load products after initialization
      const products = await iapService.getProducts();

      set({
        products,
        isInitializing: false,
      });
    } catch (error) {
      set({
        isInitializing: false,
        error: error instanceof Error ? error.message : 'Failed to initialize IAP',
      });
      throw error;
    }
  },

  loadProducts: async () => {
    const {iapService} = get();
    if (!iapService) {
      set({error: 'IAP Service not initialized'});
      return [];
    }

    set({isLoading: true, error: null});
    try {
      const products = await iapService.getProducts();
      set({products, isLoading: false});
      return products;
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to load products',
      });
      throw error;
    }
  },

  purchase: async (productId: string) => {
    const {iapService} = get();
    if (!iapService) {
      set({error: 'IAP Service not initialized'});
      throw new Error('IAP Service not initialized');
    }

    set({isPurchasing: true, error: null, currentPurchase: null});
    try {
      const purchase = await iapService.purchase(productId);

      set({
        currentPurchase: purchase,
        isPurchasing: false,
      });

      return purchase;
    } catch (error) {
      set({
        isPurchasing: false,
        error: error instanceof Error ? error.message : 'Purchase failed',
      });
      throw error;
    }
  },

  finishPurchase: async (purchase: SubscriptionPurchase) => {
    const {iapService} = get();
    if (!iapService) {
      set({error: 'IAP Service not initialized'});
      throw new Error('IAP Service not initialized');
    }

    set({isLoading: true, error: null});
    try {
      await iapService.finishPurchase(purchase);

      set({
        currentPurchase: null,
        isLoading: false,
      });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to finish purchase',
      });
      throw error;
    }
  },

  restorePurchases: async () => {
    const {iapService} = get();
    if (!iapService) {
      set({error: 'IAP Service not initialized'});
      return;
    }

    set({isLoading: true, error: null});
    try {
      await iapService.restorePurchases();

      set({isLoading: false});
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : 'Failed to restore purchases',
      });
      throw error;
    }
  },

  setCurrentPurchase: (purchase: SubscriptionPurchase | null) => set({currentPurchase: purchase}),

  clearError: () => set({error: null}),
}));

// Hooks for easier access
export const useProducts = () => useIAPStore((state) => state.products);
export const useIsPurchasing = () => useIAPStore((state) => state.isPurchasing);
export const useIAPError = () => useIAPStore((state) => state.error);
