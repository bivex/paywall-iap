import { getPlatformSettings } from "@/actions/platform-settings";

import { SettingsPageClient } from "./settings-page-client";

export default async function SettingsPage() {
  const settings = await getPlatformSettings();
  return <SettingsPageClient initialSettings={settings} />;
}
