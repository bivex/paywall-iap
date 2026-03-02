"use client";

import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useCallback } from "react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";

const ACTION_OPTIONS = [
  { value: "all", label: "All Actions" },
  { value: "grant_subscription", label: "grant_subscription" },
  { value: "revoke_subscription", label: "revoke_subscription" },
  { value: "grant_refund", label: "grant_refund" },
  { value: "cancel_subscription", label: "cancel_subscription" },
  { value: "update_plan_price", label: "update_plan_price" },
  { value: "trigger_dunning", label: "trigger_dunning" },
  { value: "replay_webhook", label: "replay_webhook" },
  { value: "manual_renewal", label: "manual_renewal" },
];

export function AuditLogFilters() {
  const router = useRouter();
  const pathname = usePathname();
  const sp = useSearchParams();

  const update = useCallback(
    (key: string, value: string) => {
      const params = new URLSearchParams(sp.toString());
      if (value && value !== "all") {
        params.set(key, value);
      } else {
        params.delete(key);
      }
      params.delete("page");
      router.push(`${pathname}?${params.toString()}`);
    },
    [router, pathname, sp]
  );

  return (
    <div className="flex flex-wrap gap-2">
      <Input
        placeholder="Search admin or target…"
        className="w-52"
        defaultValue={sp.get("search") ?? ""}
        onChange={(e) => {
          clearTimeout((window as any)._auditSearchTimer);
          (window as any)._auditSearchTimer = setTimeout(() => update("search", e.target.value), 400);
        }}
      />
      <Select value={sp.get("action") ?? "all"} onValueChange={(v) => update("action", v)}>
        <SelectTrigger className="w-48">
          <SelectValue placeholder="All Actions" />
        </SelectTrigger>
        <SelectContent>
          {ACTION_OPTIONS.map((o) => (
            <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Input
        type="date"
        className="w-40"
        defaultValue={sp.get("from") ?? ""}
        onChange={(e) =>
          update("from", e.target.value ? new Date(e.target.value).toISOString() : "")
        }
      />
      <Input
        type="date"
        className="w-40"
        defaultValue={sp.get("to") ?? ""}
        onChange={(e) =>
          update("to", e.target.value ? new Date(e.target.value + "T23:59:59Z").toISOString() : "")
        }
      />
      {(sp.get("search") || sp.get("action") || sp.get("from") || sp.get("to")) && (
        <Button variant="ghost" size="sm" onClick={() => router.push(pathname)}>
          Clear
        </Button>
      )}
    </div>
  );
}
