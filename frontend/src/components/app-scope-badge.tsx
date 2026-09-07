"use client";

import { Badge } from "@/components/ui/badge";
import { getSelectedApp, useAppStore } from "@/stores/app-store";

export function AppScopeBadge() {
  const app = useAppStore(getSelectedApp);
  if (!app) return null;
  return (
    <Badge variant="outline" className="font-normal">
      {app.display_name}
    </Badge>
  );
}
