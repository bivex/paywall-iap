import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  AlertCircle, CheckCircle2, Clock, RefreshCw,
  Webhook, Activity, AlertTriangle, XCircle, Loader2,
  ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight,
} from "lucide-react";
import Link from "next/link";
import { getRevenueOps } from "@/actions/revenue-ops";
import type { DunningRow, WebhookRow } from "@/actions/revenue-ops";
import { ReplayWebhookButton } from "./_components/replay-webhook-button";
import { WebhookTable } from "./_components/webhook-table";

/* ─── helpers ─────────────────────────────────────────── */
function fmtDate(iso: string | null) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString("en-US", {
    month: "short", day: "numeric", hour: "2-digit", minute: "2-digit",
  });
}

const PROVIDER_COLOR: Record<string, string> = {
  stripe: "bg-violet-500/10 text-violet-600 border-violet-500/20",
  apple:  "bg-blue-500/10 text-blue-600 border-blue-500/20",
  google: "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
};

const DUNNING_STATUS_COLOR: Record<string, string> = {
  pending:     "bg-yellow-500/10 text-yellow-600 border-yellow-500/20",
  in_progress: "bg-blue-500/10 text-blue-600 border-blue-500/20",
  recovered:   "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
  failed:      "bg-red-500/10 text-red-600 border-red-500/20",
};

/* ─── pagination bar ─────────────────────────────────── */
function PaginationBar({
  page, totalPages, total, pageSize, paramKey,
}: {
  page: number;
  totalPages: number;
  total: number;
  pageSize: number;
  paramKey: string;
}) {
  const from = Math.min((page - 1) * pageSize + 1, total);
  const to   = Math.min(page * pageSize, total);

  function href(p: number) {
    return `?${paramKey}=${p}#webhooks`;
  }

  return (
    <div className="flex items-center justify-between px-1 pt-3 border-t">
      <p className="text-xs text-muted-foreground">
        Showing <span className="font-medium text-foreground">{from}–{to}</span> of{" "}
        <span className="font-medium text-foreground">{total}</span>
      </p>
      <div className="flex items-center gap-1">
        <Button variant="outline" size="icon" className="h-7 w-7" disabled={page <= 1} asChild={page > 1}>
          {page > 1 ? <Link href={href(1)}><ChevronsLeft className="h-3.5 w-3.5" /></Link> : <span><ChevronsLeft className="h-3.5 w-3.5" /></span>}
        </Button>
        <Button variant="outline" size="icon" className="h-7 w-7" disabled={page <= 1} asChild={page > 1}>
          {page > 1 ? <Link href={href(page - 1)}><ChevronLeft className="h-3.5 w-3.5" /></Link> : <span><ChevronLeft className="h-3.5 w-3.5" /></span>}
        </Button>

        {/* page numbers */}
        {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
          let p: number;
          if (totalPages <= 5) {
            p = i + 1;
          } else if (page <= 3) {
            p = i + 1;
          } else if (page >= totalPages - 2) {
            p = totalPages - 4 + i;
          } else {
            p = page - 2 + i;
          }
          return (
            <Button
              key={p}
              variant={p === page ? "default" : "outline"}
              size="icon"
              className="h-7 w-7 text-xs"
              asChild={p !== page}
            >
              {p !== page ? <Link href={href(p)}>{p}</Link> : <span>{p}</span>}
            </Button>
          );
        })}

        <Button variant="outline" size="icon" className="h-7 w-7" disabled={page >= totalPages} asChild={page < totalPages}>
          {page < totalPages ? <Link href={href(page + 1)}><ChevronRight className="h-3.5 w-3.5" /></Link> : <span><ChevronRight className="h-3.5 w-3.5" /></span>}
        </Button>
        <Button variant="outline" size="icon" className="h-7 w-7" disabled={page >= totalPages} asChild={page < totalPages}>
          {page < totalPages ? <Link href={href(totalPages)}><ChevronsRight className="h-3.5 w-3.5" /></Link> : <span><ChevronsRight className="h-3.5 w-3.5" /></span>}
        </Button>
      </div>
    </div>
  );
}

/* ─── page ────────────────────────────────────────────── */
export default async function RevenueOpsPage({
  searchParams,
}: {
  searchParams: Promise<{ wh_page?: string }>;
}) {
  const sp = await searchParams;
  const whPage = Math.max(1, parseInt(sp.wh_page ?? "1", 10) || 1);

  const report = await getRevenueOps(whPage);

  if (!report) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-24 text-muted-foreground">
        <AlertCircle className="h-8 w-8" />
        <p className="text-sm">Failed to load — make sure you are logged in.</p>
      </div>
    );
  }

  const { dunning, webhooks, matomo } = report;
  const activeDunning = dunning.stats.pending + dunning.stats.in_progress;

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Revenue Ops</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Dunning queue · Webhook inbox · Matomo event pipeline · live DB data
          </p>
        </div>
        <div className="flex gap-2 shrink-0">
          {activeDunning > 0 && (
            <Badge variant="outline" className="text-xs text-amber-600 border-amber-500/40 bg-amber-500/5">
              <span className="h-1.5 w-1.5 rounded-full bg-amber-500 mr-1.5 inline-block animate-pulse" />
              {activeDunning} dunning active
            </Badge>
          )}
          {webhooks.unprocessed > 0 && (
            <Badge variant="outline" className="text-xs text-red-600 border-red-500/40 bg-red-500/5">
              <span className="h-1.5 w-1.5 rounded-full bg-red-500 mr-1.5 inline-block animate-pulse" />
              {webhooks.unprocessed} webhook{webhooks.unprocessed > 1 ? "s" : ""} pending
            </Badge>
          )}
        </div>
      </div>

      {/* Summary stat cards */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          { label: "Active Dunning",      value: activeDunning,           icon: RefreshCw,    color: activeDunning > 0 ? "text-amber-500" : "text-muted-foreground", bg: "bg-amber-500/10"   },
          { label: "Recovered",           value: dunning.stats.recovered, icon: CheckCircle2, color: "text-emerald-500",                                              bg: "bg-emerald-500/10" },
          { label: "Webhooks Unprocessed",value: webhooks.unprocessed,    icon: Webhook,      color: webhooks.unprocessed > 0 ? "text-red-500" : "text-muted-foreground", bg: "bg-red-500/10" },
          { label: "Total Webhooks",      value: webhooks.total,          icon: Activity,     color: "text-blue-500",                                                bg: "bg-blue-500/10"    },
        ].map((s) => {
          const Icon = s.icon;
          return (
            <Card key={s.label} className="py-4">
              <CardContent className="px-4 py-0 flex items-center justify-between">
                <div>
                  <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">{s.label}</p>
                  <p className={`text-2xl font-bold tabular-nums ${s.color}`}>{s.value}</p>
                </div>
                <div className={`flex h-9 w-9 items-center justify-center rounded-full ${s.bg}`}>
                  <Icon className={`h-4 w-4 ${s.color}`} />
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Tabs defaultValue="webhooks">
        <TabsList>
          <TabsTrigger value="webhooks">
            Webhooks
            {webhooks.unprocessed > 0 && (
              <span className="ml-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white">
                {webhooks.unprocessed}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="dunning">
            Dunning
            {activeDunning > 0 && (
              <span className="ml-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-amber-500 text-[10px] font-bold text-white">
                {activeDunning}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="matomo">Matomo Pipeline</TabsTrigger>
        </TabsList>

        {/* ── WEBHOOK INBOX ── */}
        <TabsContent value="webhooks" id="webhooks" className="mt-4 space-y-4">
          {/* By-provider breakdown */}
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            {webhooks.by_provider.map((p) => {
              const unproc = p.total - p.processed;
              return (
                <Card key={p.provider} className="py-3">
                  <CardContent className="px-4 py-0 flex items-center justify-between">
                    <div>
                      <Badge className={`${PROVIDER_COLOR[p.provider.toLowerCase()] ?? "bg-muted"} border text-xs capitalize mb-1`}>
                        {p.provider}
                      </Badge>
                      <p className="text-lg font-bold tabular-nums">{p.total}</p>
                      <p className="text-xs text-muted-foreground">{p.processed} processed · {unproc} pending</p>
                    </div>
                    {unproc > 0
                      ? <AlertTriangle className="h-5 w-5 text-amber-500 shrink-0" />
                      : <CheckCircle2 className="h-5 w-5 text-emerald-500 shrink-0" />
                    }
                  </CardContent>
                </Card>
              );
            })}
          </div>

          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-semibold">Recent Webhook Events</CardTitle>
                <span className="text-xs text-muted-foreground">
                  Page {webhooks.page} of {webhooks.total_pages} · {webhooks.total} total
                </span>
              </div>
            </CardHeader>
            <CardContent className="pt-0 space-y-0">
              <WebhookTable rows={webhooks.events} />
              <PaginationBar
                page={webhooks.page}
                totalPages={webhooks.total_pages}
                total={webhooks.total}
                pageSize={webhooks.page_size}
                paramKey="wh_page"
              />
            </CardContent>
          </Card>
        </TabsContent>

        {/* ── DUNNING QUEUE ── */}
        <TabsContent value="dunning" className="mt-4">
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-sm font-semibold">Active Dunning Queue</CardTitle>
                <div className="flex gap-3 text-xs text-muted-foreground">
                  <span>Pending: <span className="font-medium text-foreground">{dunning.stats.pending}</span></span>
                  <span>In Progress: <span className="font-medium text-foreground">{dunning.stats.in_progress}</span></span>
                  <span>Recovered: <span className="font-medium text-emerald-500">{dunning.stats.recovered}</span></span>
                  <span>Failed: <span className="font-medium text-red-500">{dunning.stats.failed}</span></span>
                </div>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              <DunningTable rows={dunning.queue} />
            </CardContent>
          </Card>
        </TabsContent>

        {/* ── MATOMO PIPELINE ── */}
        <TabsContent value="matomo" className="mt-4">
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {[
              { label: "Pending",    value: matomo.stats.pending,    icon: Clock,       color: "text-amber-500",  bg: "bg-amber-500/10"  },
              { label: "Processing", value: matomo.stats.processing, icon: Loader2,     color: "text-blue-500",   bg: "bg-blue-500/10"   },
              { label: "Sent",       value: matomo.stats.sent,       icon: CheckCircle2,color: "text-emerald-500",bg: "bg-emerald-500/10" },
              { label: "Failed",     value: matomo.stats.failed,     icon: XCircle,     color: "text-red-500",    bg: "bg-red-500/10"    },
            ].map((s) => {
              const Icon = s.icon;
              return (
                <Card key={s.label} className="py-4">
                  <CardContent className="px-4 py-0 flex items-center justify-between">
                    <div>
                      <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">{s.label}</p>
                      <p className={`text-2xl font-bold tabular-nums ${s.color}`}>{s.value}</p>
                    </div>
                    <div className={`flex h-9 w-9 items-center justify-center rounded-full ${s.bg}`}>
                      <Icon className={`h-4 w-4 ${s.color}`} />
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
          {matomo.stats.total === 0 && (
            <Card className="mt-4">
              <CardContent className="py-12 text-center text-sm text-muted-foreground">
                No Matomo staged events in the pipeline.
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

/* ─── sub-components ──────────────────────────────────── */
// WebhookTable is a client component in _components/webhook-table.tsx (sortable)

function DunningTable({ rows }: { rows: DunningRow[] }) {
  if (rows.length === 0) {
    return (
      <div className="py-12 text-center">
        <CheckCircle2 className="h-8 w-8 mx-auto text-emerald-500 mb-2" />
        <p className="text-sm text-muted-foreground">No active dunning — all subscriptions are healthy.</p>
      </div>
    );
  }
  return (
    <Table>
      <TableHeader>
        <TableRow className="hover:bg-transparent">
          <TableHead>User</TableHead>
          <TableHead>Plan</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Attempt</TableHead>
          <TableHead>Next Retry</TableHead>
          <TableHead>Last Attempt</TableHead>
          <TableHead>Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((d) => (
          <TableRow key={d.id}>
            <TableCell className="text-sm">{d.email}</TableCell>
            <TableCell>
              <Badge variant="secondary" className="text-xs capitalize">{d.plan_type}</Badge>
            </TableCell>
            <TableCell>
              <Badge className={`${DUNNING_STATUS_COLOR[d.status] ?? "bg-muted"} border text-xs`}>
                {d.status.replace("_", " ")}
              </Badge>
            </TableCell>
            <TableCell className="text-sm font-mono tabular-nums">
              {d.attempt_count}/{d.max_attempts}
            </TableCell>
            <TableCell className="text-xs text-muted-foreground whitespace-nowrap">{fmtDate(d.next_attempt_at)}</TableCell>
            <TableCell className="text-xs text-muted-foreground whitespace-nowrap">{fmtDate(d.last_attempt_at)}</TableCell>
            <TableCell>
              <Button variant="ghost" size="sm" asChild>
                <Link href={`/dashboard/users/${d.user_id}`}>View User →</Link>
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
