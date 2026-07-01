"use client";
import { useAppStore, getSelectedApp } from "@/stores/app-store";
import { Badge } from "@/components/ui/badge";

export function AppScopeBadge() {
  const app = useAppStore(getSelectedApp);
  if (!app) return null;
  return <Badge variant="outline" className="font-normal">{app.display_name}</Badge>;
}
