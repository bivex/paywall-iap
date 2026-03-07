import { getPricingTiers } from "@/actions/pricing";

import { PricingPageClient } from "./pricing-page-client";

export default async function PricingPage() {
  const tiers = await getPricingTiers();

  return <PricingPageClient initialTiers={tiers ?? []} loadFailed={tiers === null} />;
}
