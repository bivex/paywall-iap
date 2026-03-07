import { getExperiments } from "@/actions/experiments";

import { ExperimentsPageClient } from "./experiments-page-client";

export default async function ExperimentsPage() {
  const experiments = await getExperiments();

  return <ExperimentsPageClient initialExperiments={experiments ?? []} loadFailed={experiments === null} />;
}
