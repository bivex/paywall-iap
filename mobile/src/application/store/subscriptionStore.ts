import {create} from 'zustand';
import {persist} from 'zustand/middleware';
import {Subscription} from '../../domain/entities/Subscription';
import {AccessCheck} from '../../domain/entities/Subscription';
import {SubscriptionService} from '../../infrastructure/api/SubscriptionService';
import {useAuthStore} from './authStore';

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
          const subscriptionService = new SubscriptionService(/* apiClient */);
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
          const subscriptionService = new SubscriptionService(/* apiClient */);
          const accessCheck = await subscriptionService.checkAccess();

          set({accessCheck, isLoading: false});
          return accessCheck;
        } catch (error) {
          set({
            isLoading: false,
            error: error instanceof Error ? error.message : 'Access check failed',
          });
          throw error;
        }
      },

      updateSubscription: async (planType: string) => {
        set({isLoading: true, error: null});
        try {
          const subscriptionService = new SubscriptionService(/* apiClient */);
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
          const subscriptionService = new SubscriptionService(/* apiClient */);
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
