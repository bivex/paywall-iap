import { getWinbackCampaigns } from "@/actions/winback";

import { WinbackPageClient } from "./winback-page-client";

export default async function WinbackPage() {
  const campaigns = await getWinbackCampaigns();

  return <WinbackPageClient initialCampaigns={campaigns ?? []} loadFailed={campaigns === null} />;
}
