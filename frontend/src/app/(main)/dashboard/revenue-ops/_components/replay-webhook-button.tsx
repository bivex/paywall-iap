"use client";

import { useState, useTransition } from "react";
import { Button } from "@/components/ui/button";
import { RefreshCw, CheckCircle2, AlertCircle } from "lucide-react";
import { replayWebhook } from "@/actions/revenue-ops";

export function ReplayWebhookButton({ webhookId }: { webhookId: string }) {
  const [status, setStatus] = useState<"idle" | "ok" | "error">("idle");
  const [isPending, startTransition] = useTransition();

  function handleClick() {
    startTransition(async () => {
      const ok = await replayWebhook(webhookId);
      setStatus(ok ? "ok" : "error");
      // Reset after 3s
      setTimeout(() => setStatus("idle"), 3000);
    });
  }

  if (status === "ok") {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-emerald-600 font-medium">
        <CheckCircle2 className="h-3.5 w-3.5" /> Queued
      </span>
    );
  }
  if (status === "error") {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-red-500 font-medium">
        <AlertCircle className="h-3.5 w-3.5" /> Failed
      </span>
    );
  }

  return (
    <Button
      variant="outline"
      size="sm"
      className="h-7 text-xs"
      disabled={isPending}
      onClick={handleClick}
    >
      <RefreshCw className={`h-3 w-3 mr-1 ${isPending ? "animate-spin" : ""}`} />
      Replay
    </Button>
  );
}
