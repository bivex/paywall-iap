import Link from "next/link";

import {
  Activity,
  AlertCircle,
  AlertTriangle,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Clock,
  Loader2,
  RefreshCw,
  Webhook,
  XCircle,
} from "lucide-react";

import { getRevenueOps } from "@/actions/revenue-ops";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { DunningQueueCard, getActiveDunningCount, sortDunningRows } from "./_components/dunning-queue-card";
import { PendingWebhookTable, WebhookTable } from "./_components/webhook-table";

const PROVIDER_COLOR: Record<string, string> = {
  stripe: "bg-violet-500/10 text-violet-600 border-violet-500/20",
  apple: "bg-blue-500/10 text-blue-600 border-blue-500/20",
  google: "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
};

/* ─── pagination bar ─────────────────────────────────── */
function PaginationBar({
  page,
  totalPages,
  total,
  pageSize,
  paramKey,
  extraParams = "",
}: {
  page: number;
  totalPages: number;
  total: number;
  pageSize: number;
  paramKey: string;
  extraParams?: string;
}) {
  const from = Math.min((page - 1) * pageSize + 1, total);
  const to = Math.min(page * pageSize, total);

  function href(p: number) {
    const extra = extraParams ? `&${extraParams}` : "";
    return `?${paramKey}=${p}${extra}#webhooks`;
  }

  return (
    <div className="flex items-center justify-between border-t px-1 pt-3">
      <p className="text-muted-foreground text-xs">
        Showing{" "}
        <span className="font-medium text-foreground">
          {from}–{to}
        </span>{" "}
        of <span className="font-medium text-foreground">{total}</span>
      </p>
      <div className="flex items-center gap-1">
        <Button variant="outline" size="icon" className="h-7 w-7" disabled={page <= 1} asChild={page > 1}>
          {page > 1 ? (
            <Link href={href(1)}>
              <ChevronsLeft className="h-3.5 w-3.5" />
            </Link>
          ) : (
            <span>
              <ChevronsLeft className="h-3.5 w-3.5" />
            </span>
          )}
        </Button>
        <Button variant="outline" size="icon" className="h-7 w-7" disabled={page <= 1} asChild={page > 1}>
          {page > 1 ? (
            <Link href={href(page - 1)}>
              <ChevronLeft className="h-3.5 w-3.5" />
            </Link>
          ) : (
            <span>
              <ChevronLeft className="h-3.5 w-3.5" />
            </span>
          )}
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

        <Button
          variant="outline"
          size="icon"
          className="h-7 w-7"
          disabled={page >= totalPages}
          asChild={page < totalPages}
        >
          {page < totalPages ? (
            <Link href={href(page + 1)}>
              <ChevronRight className="h-3.5 w-3.5" />
            </Link>
          ) : (
            <span>
              <ChevronRight className="h-3.5 w-3.5" />
            </span>
          )}
        </Button>
        <Button
          variant="outline"
          size="icon"
          className="h-7 w-7"
          disabled={page >= totalPages}
          asChild={page < totalPages}
        >
          {page < totalPages ? (
            <Link href={href(totalPages)}>
              <ChevronsRight className="h-3.5 w-3.5" />
            </Link>
          ) : (
            <span>
              <ChevronsRight className="h-3.5 w-3.5" />
            </span>
          )}
        </Button>
      </div>
    </div>
  );
}

/* ─── page ────────────────────────────────────────────── */
export default async function RevenueOpsPage({
  searchParams,
}: {
  searchParams: Promise<{ wh_page?: string; wh_sort?: string; wh_pending?: string; dunning_sort?: string }>;
}) {
  const sp = await searchParams;
  const whPage = Math.max(1, parseInt(sp.wh_page ?? "1", 10) || 1);
  const whSort = sp.wh_sort as "status" | "provider" | "event_type" | "created_at" | "actions" | undefined;
  const whPending = sp.wh_pending === "1";
  const dunningSort = sp.dunning_sort ?? "date_desc"; // newest first

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
  const activeDunning = getActiveDunningCount(dunning.stats);
  const sortedDunning = sortDunningRows(dunning.queue, dunningSort);

  const buildDunningSortUrl = (s: string) => {
    const qs = new URLSearchParams();
    if (sp.wh_page && sp.wh_page !== "1") qs.set("wh_page", sp.wh_page);
    if (sp.wh_sort) qs.set("wh_sort", sp.wh_sort);
    if (sp.wh_pending) qs.set("wh_pending", sp.wh_pending);
    if (s !== "date_desc") qs.set("dunning_sort", s);
    const str = qs.toString();
    return str ? `?${str}#dunning` : "?#dunning";
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="font-semibold text-2xl tracking-tight">Revenue Ops</h1>
          <p className="mt-0.5 text-muted-foreground text-sm">
            Dunning queue · Webhook inbox · Matomo event pipeline · live DB data
          </p>
        </div>
        <div className="flex shrink-0 gap-2">
          {activeDunning > 0 && (
            <Badge variant="outline" className="border-amber-500/40 bg-amber-500/5 text-amber-600 text-xs">
              <span className="mr-1.5 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-amber-500" />
              {activeDunning} dunning active
            </Badge>
          )}
          {webhooks.unprocessed > 0 && (
            <Badge
              variant="outline"
              className="cursor-pointer border-red-500/40 bg-red-500/5 text-red-600 text-xs"
              asChild
            >
              <Link href={`?wh_page=${whPage}&wh_sort=status&wh_pending=1#webhooks`}>
                <span className="mr-1.5 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-red-500" />
                {webhooks.unprocessed} webhook{webhooks.unprocessed > 1 ? "s" : ""} pending
              </Link>
            </Badge>
          )}
        </div>
      </div>

      {/* Summary stat cards */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          {
            label: "Active Dunning",
            value: activeDunning,
            icon: RefreshCw,
            color: activeDunning > 0 ? "text-amber-500" : "text-muted-foreground",
            bg: "bg-amber-500/10",
          },
          {
            label: "Recovered",
            value: dunning.stats.recovered,
            icon: CheckCircle2,
            color: "text-emerald-500",
            bg: "bg-emerald-500/10",
          },
          {
            label: "Webhooks Unprocessed",
            value: webhooks.unprocessed,
            icon: Webhook,
            color: webhooks.unprocessed > 0 ? "text-red-500" : "text-muted-foreground",
            bg: "bg-red-500/10",
          },
          {
            label: "Total Webhooks",
            value: webhooks.total,
            icon: Activity,
            color: "text-blue-500",
            bg: "bg-blue-500/10",
          },
        ].map((s) => {
          const Icon = s.icon;
          return (
            <Card key={s.label} className="py-4">
              <CardContent className="flex items-center justify-between px-4 py-0">
                <div>
                  <p className="mb-1 font-semibold text-muted-foreground text-xs uppercase tracking-widest">
                    {s.label}
                  </p>
                  <p className={`font-bold text-2xl tabular-nums ${s.color}`}>{s.value}</p>
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
              <span className="ml-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 font-bold text-[10px] text-white">
                {webhooks.unprocessed}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="dunning">
            Dunning
            {activeDunning > 0 && (
              <span className="ml-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-amber-500 font-bold text-[10px] text-white">
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
                  <CardContent className="flex items-center justify-between px-4 py-0">
                    <div>
                      <Badge
                        className={`${PROVIDER_COLOR[p.provider.toLowerCase()] ?? "bg-muted"} mb-1 border text-xs capitalize`}
                      >
                        {p.provider}
                      </Badge>
                      <p className="font-bold text-lg tabular-nums">{p.total}</p>
                      <p className="text-muted-foreground text-xs">
                        {p.processed} processed · {unproc} pending
                      </p>
                    </div>
                    {unproc > 0 ? (
                      <AlertTriangle className="h-5 w-5 shrink-0 text-amber-500" />
                    ) : (
                      <CheckCircle2 className="h-5 w-5 shrink-0 text-emerald-500" />
                    )}
                  </CardContent>
                </Card>
              );
            })}
          </div>

          {/* Pending-only table — all unprocessed, no pagination */}
          {(webhooks.pending_events ?? []).length > 0 && (
            <Card id="pending-webhooks">
              <CardHeader className="pb-3">
                <div className="flex items-center gap-2">
                  <span className="h-2 w-2 animate-pulse rounded-full bg-amber-500" />
                  <CardTitle className="font-semibold text-amber-600 text-sm">
                    Pending Webhooks — {(webhooks.pending_events ?? []).length} unprocessed
                  </CardTitle>
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <PendingWebhookTable rows={webhooks.pending_events ?? []} />
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="font-semibold text-sm">Recent Webhook Events</CardTitle>
                <span className="text-muted-foreground text-xs">
                  Page {webhooks.page} of {webhooks.total_pages} · {webhooks.total} total
                </span>
              </div>
            </CardHeader>
            <CardContent className="space-y-0 pt-0">
              <WebhookTable rows={webhooks.events} initialSort={whSort} initialFilterPending={whPending} />
              <PaginationBar
                page={webhooks.page}
                totalPages={webhooks.total_pages}
                total={webhooks.total}
                pageSize={webhooks.page_size}
                paramKey="wh_page"
                extraParams={[whSort ? `wh_sort=${whSort}` : "", whPending ? `wh_pending=1` : ""]
                  .filter(Boolean)
                  .join("&")}
              />
            </CardContent>
          </Card>
        </TabsContent>

        {/* ── DUNNING QUEUE ── */}
        <TabsContent value="dunning" className="mt-4">
          <DunningQueueCard
            rows={sortedDunning}
            stats={dunning.stats}
            sort={dunningSort}
            buildSortUrl={buildDunningSortUrl}
          />
        </TabsContent>

        {/* ── MATOMO PIPELINE ── */}
        <TabsContent value="matomo" className="mt-4">
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {[
              {
                label: "Pending",
                value: matomo.stats.pending,
                icon: Clock,
                color: "text-amber-500",
                bg: "bg-amber-500/10",
              },
              {
                label: "Processing",
                value: matomo.stats.processing,
                icon: Loader2,
                color: "text-blue-500",
                bg: "bg-blue-500/10",
              },
              {
                label: "Sent",
                value: matomo.stats.sent,
                icon: CheckCircle2,
                color: "text-emerald-500",
                bg: "bg-emerald-500/10",
              },
              {
                label: "Failed",
                value: matomo.stats.failed,
                icon: XCircle,
                color: "text-red-500",
                bg: "bg-red-500/10",
              },
            ].map((s) => {
              const Icon = s.icon;
              return (
                <Card key={s.label} className="py-4">
                  <CardContent className="flex items-center justify-between px-4 py-0">
                    <div>
                      <p className="mb-1 font-semibold text-muted-foreground text-xs uppercase tracking-widest">
                        {s.label}
                      </p>
                      <p className={`font-bold text-2xl tabular-nums ${s.color}`}>{s.value}</p>
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
              <CardContent className="py-12 text-center text-muted-foreground text-sm">
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
