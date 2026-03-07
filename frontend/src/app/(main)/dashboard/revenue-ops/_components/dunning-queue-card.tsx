import Link from "next/link";

import type { DunningRow, DunningStats } from "@/actions/revenue-ops";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { SortHeader } from "@/components/ui/sort-header";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const DUNNING_STATUS_COLOR: Record<string, string> = {
  pending: "bg-yellow-500/10 text-yellow-600 border-yellow-500/20",
  in_progress: "bg-blue-500/10 text-blue-600 border-blue-500/20",
  recovered: "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
  failed: "bg-red-500/10 text-red-600 border-red-500/20",
};

type DunningLabels = {
  title: string;
  pending: string;
  inProgress: string;
  recovered: string;
  failed: string;
  user: string;
  plan: string;
  status: string;
  attempt: string;
  nextRetry: string;
  lastAttempt: string;
  actions: string;
  empty: string;
  viewUser: string;
};

const DEFAULT_LABELS: DunningLabels = {
  title: "Active Dunning Queue",
  pending: "Pending",
  inProgress: "In Progress",
  recovered: "Recovered",
  failed: "Failed",
  user: "User",
  plan: "Plan",
  status: "Status",
  attempt: "Attempt",
  nextRetry: "Next Retry",
  lastAttempt: "Last Attempt",
  actions: "Actions",
  empty: "No active dunning — all subscriptions are healthy.",
  viewUser: "View User",
};

function fmtDate(iso: string | null) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function getActiveDunningCount(stats: DunningStats): number {
  return stats.pending + stats.in_progress;
}

export function sortDunningRows(rows: DunningRow[], sort = "date_desc"): DunningRow[] {
  return [...rows].sort((a, b) => {
    const ta = a.next_attempt_at ? new Date(a.next_attempt_at).getTime() : 0;
    const tb = b.next_attempt_at ? new Date(b.next_attempt_at).getTime() : 0;
    return sort === "date_asc" ? ta - tb : tb - ta;
  });
}

export function DunningQueueCard({
  rows,
  stats,
  sort,
  buildSortUrl,
  labels = DEFAULT_LABELS,
}: {
  rows: DunningRow[];
  stats: DunningStats;
  sort?: string;
  buildSortUrl?: (sort: string) => string;
  labels?: DunningLabels;
}) {
  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-4">
          <CardTitle className="font-semibold text-sm">{labels.title}</CardTitle>
          <div className="flex flex-wrap gap-3 text-muted-foreground text-xs">
            <span>
              {labels.pending}: <span className="font-medium text-foreground">{stats.pending}</span>
            </span>
            <span>
              {labels.inProgress}: <span className="font-medium text-foreground">{stats.in_progress}</span>
            </span>
            <span>
              {labels.recovered}: <span className="font-medium text-emerald-500">{stats.recovered}</span>
            </span>
            <span>
              {labels.failed}: <span className="font-medium text-red-500">{stats.failed}</span>
            </span>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        {rows.length === 0 ? (
          <div className="py-12 text-center text-muted-foreground text-sm">{labels.empty}</div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>{labels.user}</TableHead>
                <TableHead>{labels.plan}</TableHead>
                <TableHead>{labels.status}</TableHead>
                <TableHead>{labels.attempt}</TableHead>
                <TableHead>
                  {sort && buildSortUrl ? (
                    <SortHeader
                      label={labels.nextRetry}
                      sortKey="date"
                      currentSort={sort}
                      ascHref={buildSortUrl("date_asc")}
                      descHref={buildSortUrl("date_desc")}
                    />
                  ) : (
                    labels.nextRetry
                  )}
                </TableHead>
                <TableHead>{labels.lastAttempt}</TableHead>
                <TableHead>{labels.actions}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="text-sm">{d.email}</TableCell>
                  <TableCell>
                    <Badge variant="secondary" className="text-xs capitalize">
                      {d.plan_type}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge className={`${DUNNING_STATUS_COLOR[d.status] ?? "bg-muted"} border text-xs`}>
                      {d.status.replace("_", " ")}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-sm tabular-nums">
                    {d.attempt_count}/{d.max_attempts}
                  </TableCell>
                  <TableCell className="whitespace-nowrap text-muted-foreground text-xs">
                    {fmtDate(d.next_attempt_at)}
                  </TableCell>
                  <TableCell className="whitespace-nowrap text-muted-foreground text-xs">
                    {fmtDate(d.last_attempt_at)}
                  </TableCell>
                  <TableCell>
                    <Button variant="ghost" size="sm" asChild>
                      <Link href={`/dashboard/users/${d.user_id}`}>{labels.viewUser} →</Link>
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}
