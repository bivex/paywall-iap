/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-02 06:27
 * Last Updated: 2026-03-02 06:27
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

"use client";

import { useRouter, useSearchParams, usePathname } from "next/navigation";
import { useTransition, useCallback } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export function SubscriptionsFilters() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [, startTransition] = useTransition();

  const updateParam = useCallback(
    (key: string, value: string) => {
      const params = new URLSearchParams(searchParams.toString());
      if (value && value !== "all") {
        params.set(key, value);
      } else {
        params.delete(key);
      }
      params.delete("page");
      startTransition(() => {
        router.push(`${pathname}?${params.toString()}`);
      });
    },
    [router, pathname, searchParams],
  );

  const sourceValue = (() => {
    const src = searchParams.get("source");
    const plat = searchParams.get("platform");
    if (src === "iap" && plat === "ios") return "apple";
    if (src === "iap" && plat === "android") return "google";
    if (src) return src;
    return "all";
  })();

  const onSourceChange = (v: string) => {
    const params = new URLSearchParams(searchParams.toString());
    params.delete("page");
    if (v === "all") {
      params.delete("source");
      params.delete("platform");
    } else if (v === "apple") {
      params.set("source", "iap");
      params.set("platform", "ios");
    } else if (v === "google") {
      params.set("source", "iap");
      params.set("platform", "android");
    } else {
      params.set("source", v);
      params.delete("platform");
    }
    startTransition(() => router.push(`${pathname}?${params.toString()}`));
  };

  return (
    <div className="flex flex-wrap gap-2">
      <Select
        value={searchParams.get("status") ?? "all"}
        onValueChange={(v) => updateParam("status", v)}
      >
        <SelectTrigger className="w-40">
          <SelectValue placeholder="Status: All" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Statuses</SelectItem>
          <SelectItem value="active">Active</SelectItem>
          <SelectItem value="grace">Grace</SelectItem>
          <SelectItem value="cancelled">Cancelled</SelectItem>
          <SelectItem value="expired">Expired</SelectItem>
        </SelectContent>
      </Select>

      <Select value={sourceValue} onValueChange={onSourceChange}>
        <SelectTrigger className="w-36">
          <SelectValue placeholder="Source: All" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Sources</SelectItem>
          <SelectItem value="apple">Apple IAP</SelectItem>
          <SelectItem value="google">Google Play</SelectItem>
          <SelectItem value="stripe">Stripe</SelectItem>
          <SelectItem value="paddle">Paddle</SelectItem>
        </SelectContent>
      </Select>

      <Select
        value={searchParams.get("plan_type") ?? "all"}
        onValueChange={(v) => updateParam("plan_type", v)}
      >
        <SelectTrigger className="w-36">
          <SelectValue placeholder="Plan: All" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Plans</SelectItem>
          <SelectItem value="monthly">Monthly</SelectItem>
          <SelectItem value="annual">Annual</SelectItem>
          <SelectItem value="lifetime">Lifetime</SelectItem>
        </SelectContent>
      </Select>

      <Input
        type="date"
        className="w-40"
        defaultValue={searchParams.get("date_from") ?? ""}
        onBlur={(e) => updateParam("date_from", e.target.value)}
      />

      <Input
        type="date"
        className="w-40"
        defaultValue={searchParams.get("date_to") ?? ""}
        onBlur={(e) => updateParam("date_to", e.target.value)}
      />

      <Input
        type="search"
        className="w-52"
        placeholder="Search email…"
        defaultValue={searchParams.get("search") ?? ""}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            updateParam("search", (e.target as HTMLInputElement).value);
          }
        }}
      />

      {Array.from(searchParams.keys()).length > 0 && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => startTransition(() => router.push(pathname))}
        >
          Clear
        </Button>
      )}
    </div>
  );
}
