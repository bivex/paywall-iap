"use client";

import { PricingTierManager } from "@/components/pricing/pricing-tier-manager";
import type { PricingTier } from "@/lib/pricing-tiers";

export function PricingPageClient({ initialTiers, loadFailed }: { initialTiers: PricingTier[]; loadFailed: boolean }) {
  return <PricingTierManager initialTiers={initialTiers} loadFailed={loadFailed} />;
}
