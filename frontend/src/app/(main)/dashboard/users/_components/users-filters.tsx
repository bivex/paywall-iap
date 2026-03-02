"use client";

import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { useCallback } from "react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";

export function UsersFilters() {
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

  const hasFilters = sp.get("search") || sp.get("platform") || sp.get("role");

  return (
    <div className="flex flex-wrap gap-2">
      <Input
        placeholder="Search email or user ID…"
        className="w-64"
        defaultValue={sp.get("search") ?? ""}
        onChange={(e) => {
          clearTimeout((window as any)._userSearchTimer);
          (window as any)._userSearchTimer = setTimeout(() => update("search", e.target.value), 400);
        }}
      />
      <Select value={sp.get("platform") ?? "all"} onValueChange={(v) => update("platform", v)}>
        <SelectTrigger className="w-36">
          <SelectValue placeholder="Platform" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Platforms</SelectItem>
          <SelectItem value="ios">iOS</SelectItem>
          <SelectItem value="android">Android</SelectItem>
          <SelectItem value="web">Web</SelectItem>
        </SelectContent>
      </Select>
      <Select value={sp.get("role") ?? "all"} onValueChange={(v) => update("role", v)}>
        <SelectTrigger className="w-32">
          <SelectValue placeholder="Role" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Roles</SelectItem>
          <SelectItem value="user">User</SelectItem>
          <SelectItem value="admin">Admin</SelectItem>
          <SelectItem value="superadmin">Superadmin</SelectItem>
        </SelectContent>
      </Select>
      {hasFilters && (
        <Button variant="ghost" size="sm" onClick={() => router.push(pathname)}>
          Clear
        </Button>
      )}
    </div>
  );
}
