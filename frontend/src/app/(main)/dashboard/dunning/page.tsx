import Link from "next/link";

import { AlertCircle, CheckCircle2, Clock, RefreshCw, XCircle } from "lucide-react";
import { getTranslations } from "next-intl/server";

import { getRevenueOps } from "@/actions/revenue-ops";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

import {
  DunningQueueCard,
  getActiveDunningCount,
  sortDunningRows,
} from "../revenue-ops/_components/dunning-queue-card";

const DUNNING_STATUSES = ["all", "pending", "in_progress", "recovered", "failed"] as const;

type DunningStatusFilter = (typeof DUNNING_STATUSES)[number];

function isDunningStatusFilter(value?: string): value is DunningStatusFilter {
  return Boolean(value && DUNNING_STATUSES.includes(value as DunningStatusFilter));
}

export default async function DunningPage({
  searchParams,
}: {
  searchParams: Promise<{ dunning_sort?: string; q?: string; status?: string }>;
}) {
  const sp = await searchParams;
  const t = await getTranslations("dunning");
  const dunningSort = sp.dunning_sort ?? "date_desc";
  const query = sp.q?.trim() ?? "";
  const statusFilter = isDunningStatusFilter(sp.status) ? sp.status : "all";
  const report = await getRevenueOps(1);

  if (!report) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-24 text-muted-foreground">
        <AlertCircle className="h-8 w-8" />
        <p className="text-sm">{t("states.loadFailed")}</p>
      </div>
    );
  }

  const { dunning } = report;
  const activeDunning = getActiveDunningCount(dunning.stats);
  const filteredDunning = dunning.queue.filter((row) => {
    const matchesStatus = statusFilter === "all" || row.status === statusFilter;
    const haystack = `${row.email} ${row.user_id} ${row.plan_type}`.toLowerCase();
    const matchesQuery = query.length === 0 || haystack.includes(query.toLowerCase());
    return matchesStatus && matchesQuery;
  });
  const sortedDunning = sortDunningRows(filteredDunning, dunningSort);
  const hasFilters = query.length > 0 || statusFilter !== "all";

  const buildDunningUrl = (params: { sort?: string; q?: string; status?: DunningStatusFilter }) => {
    const qs = new URLSearchParams();
    if (params.q) qs.set("q", params.q);
    if (params.status && params.status !== "all") qs.set("status", params.status);
    if (params.sort && params.sort !== "date_desc") qs.set("dunning_sort", params.sort);
    const str = qs.toString();
    return str ? `/dashboard/dunning?${str}` : "/dashboard/dunning";
  };

  const buildDunningSortUrl = (sort: string) => buildDunningUrl({ sort, q: query, status: statusFilter });
  const resetFiltersUrl = buildDunningUrl({ sort: dunningSort });

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="font-semibold text-2xl tracking-tight">{t("title")}</h1>
          <p className="mt-0.5 text-muted-foreground text-sm">{t("subtitle")}</p>
        </div>
        <div className="flex shrink-0 gap-2">
          {activeDunning > 0 && (
            <div className="rounded-md border border-amber-500/40 bg-amber-500/5 px-3 py-1.5 text-amber-600 text-xs">
              <span className="mr-1.5 inline-block h-1.5 w-1.5 animate-pulse rounded-full bg-amber-500" />
              {activeDunning} {t("badges.active")}
            </div>
          )}
          <Button size="sm" variant="outline" asChild>
            <Link href="/dashboard/revenue-ops#dunning">{t("actions.openRevenueOps")}</Link>
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          {
            label: t("summary.active"),
            value: activeDunning,
            icon: RefreshCw,
            color: activeDunning > 0 ? "text-amber-500" : "text-muted-foreground",
            bg: "bg-amber-500/10",
          },
          {
            label: t("summary.pending"),
            value: dunning.stats.pending,
            icon: Clock,
            color: "text-yellow-600",
            bg: "bg-yellow-500/10",
          },
          {
            label: t("summary.recovered"),
            value: dunning.stats.recovered,
            icon: CheckCircle2,
            color: "text-emerald-500",
            bg: "bg-emerald-500/10",
          },
          {
            label: t("summary.failed"),
            value: dunning.stats.failed,
            icon: XCircle,
            color: "text-red-500",
            bg: "bg-red-500/10",
          },
        ].map((item) => {
          const Icon = item.icon;
          return (
            <Card key={item.label} className="py-4">
              <CardContent className="flex items-center justify-between px-4 py-0">
                <div>
                  <p className="mb-1 font-semibold text-muted-foreground text-xs uppercase tracking-widest">
                    {item.label}
                  </p>
                  <p className={`font-bold text-2xl tabular-nums ${item.color}`}>{item.value}</p>
                </div>
                <div className={`flex h-9 w-9 items-center justify-center rounded-full ${item.bg}`}>
                  <Icon className={`h-4 w-4 ${item.color}`} />
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Card className="py-4">
        <CardContent className="px-4 py-0">
          <form className="flex flex-col gap-3 sm:flex-row sm:items-center" method="get">
            {dunningSort !== "date_desc" && <input type="hidden" name="dunning_sort" value={dunningSort} />}
            <div className="flex-1">
              <Input
                aria-label={t("filters.searchPlaceholder")}
                defaultValue={query}
                name="q"
                placeholder={t("filters.searchPlaceholder")}
              />
            </div>
            <select
              aria-label={t("filters.statusPlaceholder")}
              className="h-9 min-w-48 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 dark:bg-input/30"
              defaultValue={statusFilter}
              name="status"
            >
              <option value="all">{t("filters.statusAll")}</option>
              <option value="pending">{t("filters.statusPending")}</option>
              <option value="in_progress">{t("filters.statusInProgress")}</option>
              <option value="recovered">{t("filters.statusRecovered")}</option>
              <option value="failed">{t("filters.statusFailed")}</option>
            </select>
            <div className="flex gap-2">
              <Button type="submit">{t("actions.applyFilters")}</Button>
              {hasFilters && (
                <Button variant="outline" asChild>
                  <Link href={resetFiltersUrl}>{t("actions.resetFilters")}</Link>
                </Button>
              )}
            </div>
          </form>
        </CardContent>
      </Card>

      <DunningQueueCard
        rows={sortedDunning}
        stats={dunning.stats}
        sort={dunningSort}
        buildSortUrl={buildDunningSortUrl}
        labels={{
          title: t("queue.title"),
          pending: t("queue.pending"),
          inProgress: t("queue.inProgress"),
          recovered: t("queue.recovered"),
          failed: t("queue.failed"),
          user: t("queue.table.user"),
          plan: t("queue.table.plan"),
          status: t("queue.table.status"),
          attempt: t("queue.table.attempt"),
          nextRetry: t("queue.table.nextRetry"),
          lastAttempt: t("queue.table.lastAttempt"),
          actions: t("queue.table.actions"),
          empty: hasFilters ? t("queue.emptyFiltered") : t("queue.empty"),
          viewUser: t("queue.viewUser"),
        }}
      />
    </div>
  );
}
