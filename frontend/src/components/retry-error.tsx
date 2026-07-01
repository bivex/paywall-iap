"use client";

import { AlertCircle, RefreshCw } from "lucide-react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";

/**
 * Recoverable load-failure state for server-rendered pages.
 *
 * Shown when a page's data fetch failed for a non-auth reason (5xx, network) —
 * i.e. the user IS logged in. `router.refresh()` re-runs the Server Component
 * to retry the request. Auth failures (no/expired token) are handled earlier by
 * serverFetch redirecting to the login page.
 */
export function RetryError({ message = "Failed to load data." }: { message?: string }) {
  const router = useRouter();
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-24 text-muted-foreground">
      <AlertCircle className="h-8 w-8" />
      <p className="text-sm">{message}</p>
      <Button variant="outline" size="sm" onClick={() => router.refresh()}>
        <RefreshCw className="h-4 w-4 mr-1.5" />
        Retry
      </Button>
    </div>
  );
}
