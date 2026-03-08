import { getPricingTiers } from "@/actions/pricing";

import { PaywallCreatorPageClient } from "./paywall-creator-page-client";

export default async function PaywallCreatorPage() {
  const tiers = await getPricingTiers();

  return <PaywallCreatorPageClient initialTiers={tiers ?? []} loadFailed={tiers === null} />;
}
