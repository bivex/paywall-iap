import { getMultiObjectiveDashboardFromCookies } from "@/lib/server/multi-objective-admin";

import { MultiObjectivePageClient } from "./multi-objective-page-client";

export default async function MultiObjectivePage() {
  const data = await getMultiObjectiveDashboardFromCookies();

  return (
    <MultiObjectivePageClient
      initialExperiments={data.experiments}
      initialSelectedExperimentId={data.selectedExperimentId}
      initialSnapshot={data.snapshot}
      loadFailed={data.loadFailed}
    />
  );
}
