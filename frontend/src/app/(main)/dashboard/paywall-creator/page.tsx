import { getPricingTiers } from "@/actions/pricing";
import { getPaywalls } from "@/actions/paywalls";
import { PaywallCreatorPageClient } from "./paywall-creator-page-client";

export default async function PaywallCreatorPage() {
  const [tiers, paywallsRes] = await Promise.all([
    getPricingTiers(),
    getPaywalls(),
  ]);

  return (
    <PaywallCreatorPageClient
      initialTiers={tiers ?? []}
      loadFailed={tiers === null}
      initialPaywalls={paywallsRes?.paywalls ?? []}
    />
  );
}
