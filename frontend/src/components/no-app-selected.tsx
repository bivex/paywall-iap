"use client";
import { Button } from "@/components/ui/button";
import { Smartphone } from "lucide-react";

export function NoAppSelected() {
  return (
    <div className="flex flex-col items-center justify-center py-16 gap-3 text-center">
      <Smartphone className="size-8 text-muted-foreground" />
      <p className="text-sm text-muted-foreground">No app selected. Choose an app from the sidebar.</p>
      <Button asChild variant="outline" size="sm">
        <a href="/dashboard/apps">Manage Apps</a>
      </Button>
    </div>
  );
}
