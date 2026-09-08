import {ApiClient, ApiResponse} from './ApiClient';
import {Subscription, AccessCheck} from '../../domain/entities/Subscription';

export interface VerifyIAPRequest {
  platform: 'ios' | 'android';
  receipt_data: string;
  product_id: string;
  transaction_id?: string;
}

export interface VerifyIAPResponse {
  subscription_id: string;
  status: string;
  expires_at: string;
  auto_renew: boolean;
  plan_type: string;
  is_new: boolean;
}

export class SubscriptionService {
  constructor(private api: ApiClient) {}

  async getSubscription(): Promise<Subscription> {
    const response = await this.api.get<ApiResponse<any>>('/subscription');
    const data = response.data;
    if (!data) return data;
    return {
      id: data.id,
      userId: data.user_id || data.userId,
      status: data.status,
      source: data.source,
      platform: data.platform,
      productId: data.product_id || data.productId,
      planType: data.plan_type || data.planType || 'annual',
      expiresAt: data.expires_at || data.expiresAt,
      autoRenew: data.auto_renew ?? data.autoRenew ?? true,
      createdAt: data.created_at || data.createdAt,
      updatedAt: data.updated_at || data.updatedAt,
    };
  }

  async checkAccess(): Promise<AccessCheck> {
    const response = await this.api.get<ApiResponse<any>>('/subscription/access');
    const data = response.data;
    if (!data) return { hasAccess: false };
    return {
      hasAccess: data.has_access ?? data.hasAccess ?? false,
      expiresAt: data.expires_at ?? data.expiresAt,
      reason: data.reason,
    };
  }

  async verifyIAP(platform: 'ios' | 'android', receiptData: string, productId: string): Promise<VerifyIAPResponse> {
    const request: VerifyIAPRequest = {
      platform,
      receipt_data: receiptData,
      product_id: productId,
    };

    const response = await this.api.post<ApiResponse<VerifyIAPResponse>>('/verify/iap', request);
    return response.data;
  }

  async cancelSubscription(): Promise<void> {
    await this.api.delete<void>('/subscription');
  }

  async restoreSubscription(): Promise<Subscription> {
    const response = await this.api.post<ApiResponse<any>>('/subscription/restore', {});
    const data = response.data;
    if (!data) return data;
    return {
      id: data.id,
      userId: data.user_id || data.userId,
      status: data.status,
      source: data.source,
      platform: data.platform,
      productId: data.product_id || data.productId,
      planType: data.plan_type || data.planType || 'annual',
      expiresAt: data.expires_at || data.expiresAt,
      autoRenew: data.auto_renew ?? data.autoRenew ?? true,
      createdAt: data.created_at || data.createdAt,
      updatedAt: data.updated_at || data.updatedAt,
    };
  }
}
