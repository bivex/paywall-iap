export enum SubscriptionStatus {
  Active = 'active',
  Expired = 'expired',
  Cancelled = 'cancelled',
  Grace = 'grace',
}

export enum SubscriptionSource {
  IAP = 'iap',
  Stripe = 'stripe',
  Paddle = 'paddle',
}

export enum PlanType {
  Monthly = 'monthly',
  Annual = 'annual',
  Lifetime = 'lifetime',
}

export interface Subscription {
  id: string;
  userId: string;
  status: SubscriptionStatus;
  source: SubscriptionSource;
  platform: string;
  productId: string;
  planType: PlanType;
  expiresAt: string;
  autoRenew: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface AccessCheck {
  hasAccess: boolean;
  expiresAt?: string;
  reason?: string;
}
