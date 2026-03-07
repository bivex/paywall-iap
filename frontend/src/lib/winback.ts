export type WinbackDiscountType = "percentage" | "fixed";

export interface WinbackCampaign {
  campaign_id: string;
  discount_type: WinbackDiscountType;
  discount_value: number;
  total_offers: number;
  active_offers: number;
  accepted_offers: number;
  expired_offers: number;
  declined_offers: number;
  launched_at: string;
  latest_expiry_at: string;
}

export interface LaunchWinbackCampaignInput {
  campaign_id: string;
  discount_type: WinbackDiscountType;
  discount_value: number;
  duration_days: number;
  days_since_churn: number;
}

export const EMPTY_WINBACK_CAMPAIGN_INPUT: LaunchWinbackCampaignInput = {
  campaign_id: "",
  discount_type: "percentage",
  discount_value: 25,
  duration_days: 14,
  days_since_churn: 30,
};
