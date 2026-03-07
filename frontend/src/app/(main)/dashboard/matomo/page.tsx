import { getPlatformSettings } from "@/actions/platform-settings";

import { MatomoPageClient } from "./matomo-page-client";

export default async function MatomoPage() {
  const settings = await getPlatformSettings();

  return (
    <MatomoPageClient
      config={{
        url: settings.integrations.matomo_url,
        siteId: settings.integrations.matomo_site_id,
        hasAuthToken: Boolean(settings.integrations.matomo_auth_token),
      }}
    />
  );
}
