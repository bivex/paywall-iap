"use client";

import { AppScopeBadge } from "@/components/app-scope-badge";
import { NoAppSelected } from "@/components/no-app-selected";
import { PricingTierManager } from "@/components/pricing/pricing-tier-manager";
import type { PricingTier } from "@/lib/pricing-tiers";
import { useAppStore } from "@/stores/app-store";

export function PricingPageClient({ initialTiers, loadFailed }: { initialTiers: PricingTier[]; loadFailed: boolean }) {
  const selectedAppId = useAppStore((s) => s.selectedAppId);
  if (!selectedAppId) return <NoAppSelected />;
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-2">
        <h1 className="font-semibold text-2xl tracking-tight">Pricing</h1>
        <AppScopeBadge />
      </div>
      <PricingTierManager initialTiers={initialTiers} loadFailed={loadFailed} />
    </div>
  );
}
