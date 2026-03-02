"use client";

import Link from "next/link";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { cn } from "@/lib/utils";

interface SortHeaderProps {
  label: string;
  sortKey: string;
  currentSort: string | undefined;
  /** Pre-computed hrefs — pass from server component to avoid serialising functions */
  ascHref: string;
  descHref: string;
  className?: string;
}

/** Clickable table header that toggles asc/desc sort via URL param. */
export function SortHeader({ label, sortKey, currentSort, ascHref, descHref, className }: SortHeaderProps) {
  const asc  = `${sortKey}_asc`;
  const desc = `${sortKey}_desc`;
  const isAsc  = currentSort === asc;
  const isDesc = currentSort === desc;
  const active = isAsc || isDesc;

  const nextHref = isDesc ? ascHref : descHref;

  return (
    <Link
      href={nextHref}
      className={cn(
        "inline-flex items-center gap-1 hover:text-foreground transition-colors select-none",
        active ? "text-foreground font-semibold" : "text-muted-foreground",
        className,
      )}
    >
      {label}
      {isDesc ? (
        <ArrowDown className="h-3.5 w-3.5" />
      ) : isAsc ? (
        <ArrowUp className="h-3.5 w-3.5" />
      ) : (
        <ArrowUpDown className="h-3.5 w-3.5 opacity-50" />
      )}
    </Link>
  );
}
