import {create} from 'zustand';
import {persist} from 'zustand/middleware';
import {Subscription} from '../../domain/entities/Subscription';
import {AccessCheck} from '../../domain/entities/Subscription';
import {SubscriptionService} from '../../infrastructure/api/SubscriptionService';
import {ApiClient} from '../../infrastructure/api/ApiClient';
import {getSubscriptionService} from '../services/Services';
import {useAuthStore} from './authStore';

const getService = (): SubscriptionService => {
  try {
    return getSubscriptionService();
  } catch {
    return new SubscriptionService(new ApiClient());
  }
};

interface SubscriptionState {
  subscription: Subscription | null;
  accessCheck: AccessCheck | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchSubscription: () => Promise<void>;
  checkAccess: (featureId: string) => Promise<AccessCheck>;
  updateSubscription: (planType: string) => Promise<void>;
  cancelSubscription: (reason?: string) => Promise<void>;
  setSubscription: (subscription: Subscription | null) => void;
  clearError: () => void;
}

export const useSubscriptionStore = create<SubscriptionState>()(
  persist(
    (set, get) => ({
      subscription: null,
      accessCheck: null,
      isLoading: false,
      error: null,

      fetchSubscription: async () => {
        const {isAuthenticated} = useAuthStore.getState();
        if (!isAuthenticated) {
          set({subscription: null, error: 'Not authenticated'});
          return;
        }

        set({isLoading: true, error: null});
        try {
          const subscriptionService = getService();
          const subscription = await subscriptionService.getSubscription();

          set({subscription, isLoading: false});
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Failed to fetch subscription',
          });
        }
      },

      checkAccess: async (featureId: string) => {
        const {isAuthenticated} = useAuthStore.getState();
        if (!isAuthenticated) {
          return {
            hasAccess: false,
            reason: 'not_authenticated',
            gateId: null,
            requiredPlanType: null,
          };
        }

        set({isLoading: true, error: null});
        try {
          const subscriptionService = getService();
          const accessCheck = await subscriptionService.checkAccess();

          set({accessCheck, isLoading: false});
          return accessCheck;
        } catch (error: any) {
          const fallbackCheck: AccessCheck = {
            hasAccess: false,
            reason: error?.message || 'access_check_failed',
            gateId: null,
            requiredPlanType: null,
          };
          set({
            accessCheck: fallbackCheck,
            isLoading: false,
            error: error instanceof Error ? error.message : 'Access check failed',
          });
          return fallbackCheck;
        }
      },

      updateSubscription: async (planType: string) => {
        set({isLoading: true, error: null});
        try {
          const subscriptionService = getService();
          await subscriptionService.updateSubscriptionPlan(planType);

          // Refetch subscription
          await get().fetchSubscription();
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Failed to update subscription',
          });
          throw error;
        }
      },

      cancelSubscription: async (reason?: string) => {
        set({isLoading: true, error: null});
        try {
          const subscriptionService = getService();
          await subscriptionService.cancelSubscription();

          // Refetch subscription
          await get().fetchSubscription();
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Failed to cancel subscription',
          });
          throw error;
        }
      },

      setSubscription: (subscription: Subscription | null) => set({subscription}),

      clearError: () => set({error: null}),
    }),
    {
      name: 'subscription-storage',
      partialize: (state) => ({subscription: state.subscription}),
    },
  ),
);
