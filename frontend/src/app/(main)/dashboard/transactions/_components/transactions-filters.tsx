"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useCallback } from "react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export function TransactionsFilters() {
  const router = useRouter();
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
      router.push(`?${params.toString()}`);
    },
    [router, sp],
  );

  return (
    <div className="flex flex-wrap gap-2">
      <Select value={sp.get("status") ?? "all"} onValueChange={(v) => update("status", v)}>
        <SelectTrigger className="w-36">
          <SelectValue placeholder="All Statuses" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Statuses</SelectItem>
          <SelectItem value="success">Success</SelectItem>
          <SelectItem value="failed">Failed</SelectItem>
          <SelectItem value="refunded">Refunded</SelectItem>
        </SelectContent>
      </Select>

      <Select value={sp.get("source") ?? "all"} onValueChange={(v) => update("source", v)}>
        <SelectTrigger className="w-36">
          <SelectValue placeholder="All Sources" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Sources</SelectItem>
          <SelectItem value="iap">IAP</SelectItem>
          <SelectItem value="stripe">Stripe</SelectItem>
        </SelectContent>
      </Select>

      <Select value={sp.get("platform") ?? "all"} onValueChange={(v) => update("platform", v)}>
        <SelectTrigger className="w-32">
          <SelectValue placeholder="All Platforms" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Platforms</SelectItem>
          <SelectItem value="ios">iOS</SelectItem>
          <SelectItem value="android">Android</SelectItem>
          <SelectItem value="web">Web</SelectItem>
        </SelectContent>
      </Select>

      <Input
        type="date"
        className="w-40"
        value={sp.get("date_from") ?? ""}
        onChange={(e) => update("date_from", e.target.value)}
      />
      <Input
        type="date"
        className="w-40"
        value={sp.get("date_to") ?? ""}
        onChange={(e) => update("date_to", e.target.value)}
      />

      <Input
        placeholder="Search email…"
        className="w-52"
        defaultValue={sp.get("search") ?? ""}
        onKeyDown={(e) => {
          if (e.key === "Enter") update("search", (e.target as HTMLInputElement).value);
        }}
        onBlur={(e) => update("search", e.target.value)}
      />
    </div>
  );
}
