"use client";

import { useEffect } from "react";
import { Check, ChevronsUpDown, Smartphone } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { type App, useAppStore } from "@/stores/app-store";

export function AppSelector() {
  const { apps, selectedAppId, setApps, setSelectedAppId } = useAppStore();

  useEffect(() => {
    fetch("/api/admin/apps")
      .then((r) => r.json())
      .then((body: { apps?: App[] }) => {
        const list = body?.apps ?? [];
        setApps(list);
        // Auto-select first active app if nothing selected
        if (!selectedAppId && list.length > 0) {
          const first = list.find((a) => a.is_active) ?? list[0];
          setSelectedAppId(first.id);
        }
      })
      .catch(() => {});
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const selected = apps.find((a) => a.id === selectedAppId);

  if (apps.length === 0) return null;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="w-full justify-between gap-2 truncate">
          <Smartphone className="size-4 shrink-0 text-muted-foreground" />
          <span className="truncate text-left">
            {selected ? (selected.display_name || selected.name) : "Select app"}
          </span>
          <ChevronsUpDown className="size-3.5 shrink-0 opacity-50" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="w-56">
        <DropdownMenuLabel className="text-xs text-muted-foreground">Switch app</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {apps.map((app) => (
          <DropdownMenuItem
            key={app.id}
            onSelect={() => setSelectedAppId(app.id)}
            className="flex items-center gap-2"
          >
            <Check
              className={`size-3.5 shrink-0 transition-opacity ${app.id === selectedAppId ? "opacity-100" : "opacity-0"}`}
            />
            <span className="flex flex-col min-w-0">
              <span className="truncate font-medium">{app.display_name || app.name}</span>
              <span className="truncate text-xs text-muted-foreground">{app.bundle_id}</span>
            </span>
            {!app.is_active && (
              <span className="ml-auto text-xs text-muted-foreground">inactive</span>
            )}
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <a href="/dashboard/apps" className="text-xs text-muted-foreground">
            Manage apps…
          </a>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
