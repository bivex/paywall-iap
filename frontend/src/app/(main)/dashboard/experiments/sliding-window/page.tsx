import { getSlidingWindowDashboardFromCookies } from "@/lib/server/sliding-window-admin";

import { SlidingWindowPageClient } from "./sliding-window-page-client";

export default async function SlidingWindowPage() {
  const data = await getSlidingWindowDashboardFromCookies();

  return (
    <SlidingWindowPageClient
      initialExperiments={data.experiments}
      initialSelectedExperimentId={data.selectedExperimentId}
      initialSnapshot={data.snapshot}
      loadFailed={data.loadFailed}
    />
  );
}
