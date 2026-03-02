"use client";

import { useState, useMemo } from "react";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { CheckCircle2, Clock, ArrowUpDown, ArrowUp, ArrowDown } from "lucide-react";
import type { WebhookRow } from "@/actions/revenue-ops";
import { ReplayWebhookButton } from "./replay-webhook-button";

const PROVIDER_COLOR: Record<string, string> = {
  stripe: "bg-violet-500/10 text-violet-600 border-violet-500/20",
  apple:  "bg-blue-500/10 text-blue-600 border-blue-500/20",
  google: "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
};

function fmtDate(iso: string | null) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString("en-US", {
    month: "short", day: "numeric", hour: "2-digit", minute: "2-digit",
  });
}

type SortKey = "status" | "provider" | "event_type" | "created_at";
type SortDir = "asc" | "desc";

// Defined at module level — stable reference, no remounting issues
function SortIndicator({ active, dir }: { active: boolean; dir: SortDir }) {
  if (!active) return <ArrowUpDown className="ml-1 h-3 w-3 opacity-30 inline-block" />;
  return dir === "asc"
    ? <ArrowUp   className="ml-1 h-3 w-3 inline-block" />
    : <ArrowDown className="ml-1 h-3 w-3 inline-block" />;
}

export function WebhookTable({ rows }: { rows: WebhookRow[] }) {
  // default: pending first
  const [sortKey, setSortKey] = useState<SortKey>("status");
  const [sortDir, setSortDir] = useState<SortDir>("asc");

  function handleSort(key: SortKey) {
    if (sortKey === key) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortKey(key);
      setSortDir("asc");
    }
  }

  const sorted = useMemo(() => {
    const copy = [...rows];
    copy.sort((a, b) => {
      let cmp = 0;
      switch (sortKey) {
        case "status":
          cmp = Number(a.processed) - Number(b.processed);
          break;
        case "provider":
          cmp = a.provider.localeCompare(b.provider);
          break;
        case "event_type":
          cmp = a.event_type.localeCompare(b.event_type);
          break;
        case "created_at":
          cmp = new Date(a.created_at ?? 0).getTime() - new Date(b.created_at ?? 0).getTime();
          break;
      }
      return sortDir === "asc" ? cmp : -cmp;
    });
    return copy;
  }, [rows, sortKey, sortDir]);

  if (rows.length === 0) {
    return <p className="py-8 text-center text-sm text-muted-foreground">No webhook events found.</p>;
  }

  const thClass = "cursor-pointer select-none hover:text-foreground transition-colors";

  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead className={thClass} onClick={() => handleSort("provider")}>
            Provider <SortIndicator active={sortKey === "provider"} dir={sortDir} />
          </TableHead>
          <TableHead className={thClass} onClick={() => handleSort("event_type")}>
            Event Type <SortIndicator active={sortKey === "event_type"} dir={sortDir} />
          </TableHead>
          <TableHead>Event ID</TableHead>
          <TableHead className={thClass} onClick={() => handleSort("created_at")}>
            Received <SortIndicator active={sortKey === "created_at"} dir={sortDir} />
          </TableHead>
          <TableHead className={thClass} onClick={() => handleSort("status")}>
            Status <SortIndicator active={sortKey === "status"} dir={sortDir} />
          </TableHead>
          <TableHead className="w-20">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {sorted.map((w) => (
          <TableRow key={w.id}>
            <TableCell>
              <Badge className={`${PROVIDER_COLOR[w.provider.toLowerCase()] ?? "bg-muted text-foreground"} border text-xs capitalize`}>
                {w.provider}
              </Badge>
            </TableCell>
            <TableCell className="font-mono text-xs">{w.event_type}</TableCell>
            <TableCell className="font-mono text-xs text-muted-foreground max-w-[140px] truncate">{w.event_id}</TableCell>
            <TableCell className="text-xs text-muted-foreground whitespace-nowrap">{fmtDate(w.created_at)}</TableCell>
            <TableCell>
              {w.processed ? (
                <Badge variant="outline" className="text-xs border-emerald-500/30 text-emerald-600 bg-emerald-500/5">
                  <CheckCircle2 className="h-3 w-3 mr-1" /> processed
                </Badge>
              ) : (
                <Badge variant="outline" className="text-xs border-amber-500/30 text-amber-600 bg-amber-500/5 animate-pulse">
                  <Clock className="h-3 w-3 mr-1" /> pending
                </Badge>
              )}
            </TableCell>
            <TableCell>
              {!w.processed && <ReplayWebhookButton webhookId={w.id} />}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
