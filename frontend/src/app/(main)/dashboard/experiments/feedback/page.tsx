import { getDelayedFeedbackDashboardFromCookies } from "@/lib/server/delayed-feedback-admin";

import { DelayedFeedbackPageClient } from "./feedback-page-client";

export default async function FeedbackPage() {
  const data = await getDelayedFeedbackDashboardFromCookies();

  return (
    <DelayedFeedbackPageClient
      initialExperiments={data.experiments}
      initialSelectedExperimentId={data.selectedExperimentId}
      initialSnapshot={data.snapshot}
      loadFailed={data.loadFailed}
    />
  );
}
