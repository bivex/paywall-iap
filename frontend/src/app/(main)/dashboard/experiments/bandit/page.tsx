import { getBanditDashboardFromCookies } from "@/lib/server/bandit-admin";

import { BanditPageClient } from "./bandit-page-client";

export default async function BanditPage() {
  const data = await getBanditDashboardFromCookies();

  return (
    <BanditPageClient
      initialExperiments={data.experiments}
      initialSelectedExperimentId={data.selectedExperimentId}
      initialSnapshot={data.snapshot}
      loadFailed={data.loadFailed}
    />
  );
}
