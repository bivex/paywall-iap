export interface PricingTier {
  id: string;
  name: string;
  description: string;
  monthly_price: number | null;
  annual_price: number | null;
  currency: string;
  features: string[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PricingTierInput {
  name: string;
  description: string;
  monthly_price: number | null;
  annual_price: number | null;
  currency: string;
  features: string[];
  is_active: boolean;
}

export const EMPTY_PRICING_TIER_INPUT: PricingTierInput = {
  name: "",
  description: "",
  monthly_price: null,
  annual_price: null,
  currency: "USD",
  features: [],
  is_active: true,
};
