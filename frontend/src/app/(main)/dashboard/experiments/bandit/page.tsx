import { getBanditDashboardData } from "@/actions/bandit";

import { BanditPageClient } from "./bandit-page-client";

export default async function BanditPage() {
  const data = await getBanditDashboardData();

  return (
    <BanditPageClient
      initialExperiments={data.experiments}
      initialSelectedExperimentId={data.selectedExperimentId}
      initialSnapshot={data.snapshot}
      loadFailed={data.loadFailed}
    />
  );
}
