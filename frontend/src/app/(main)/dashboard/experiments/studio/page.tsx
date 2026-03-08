import { getStudioDashboardFromCookies } from "@/lib/server/studio-admin";

import { StudioPageClient } from "./studio-page-client";

export default async function ExperimentStudioPage() {
  const data = await getStudioDashboardFromCookies();

  return (
    <StudioPageClient
      initialExperiments={data.experiments}
      initialSelectedExperimentId={data.selectedExperimentId}
      initialSnapshot={data.snapshot}
      initialPricingTiers={data.pricingTiers}
      pricingLoadFailed={data.pricingLoadFailed}
      loadFailed={data.loadFailed}
    />
  );
}
