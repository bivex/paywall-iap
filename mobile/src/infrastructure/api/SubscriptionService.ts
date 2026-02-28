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
    const response = await this.api.get<ApiResponse<Subscription>>('/subscription');
    return response.data;
  }

  async checkAccess(): Promise<AccessCheck> {
    const response = await this.api.get<ApiResponse<AccessCheck>>('/subscription/access');
    return response.data;
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
}
