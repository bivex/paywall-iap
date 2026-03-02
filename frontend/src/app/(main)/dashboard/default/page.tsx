import { getTranslations } from "next-intl/server";
import { AlertTriangle, CheckCircle2, TrendingUp } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { getDashboardMetrics } from "@/actions/dashboard";

import { MrrTrendChart } from "./_components/mrr-trend-chart";
import { SubStatusChart } from "./_components/sub-status-chart";
import { AuditLogTable } from "./_components/audit-log-table";

// Fallback values when the API is unavailable (e.g. first load before backend starts)
const FALLBACK = {
  active_users: 0,
  active_subs: 0,
  mrr: 0,
  arr: 0,
  churn_risk: 0,
  mrr_trend: [] as import("@/actions/dashboard").MonthlyMRR[],
  status_counts: { Active: 0, Grace: 0, Cancelled: 0, Expired: 0 },
  audit_log: [] as import("@/actions/dashboard").AuditLogEntry[],
  webhook_health: [],
  last_updated: new Date().toISOString(),
} as const;

function fmt(n: number) {
  return n.toLocaleString("en-US");
}

function fmtUSD(n: number) {
  return `$${n.toLocaleString("en-US", { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`;
}

export default async function DashboardPage() {
  const [t, metrics] = await Promise.all([
    getTranslations("dashboard"),
    getDashboardMetrics(),
  ]);

  const d = metrics ?? FALLBACK;

  const kpiCards = [
    {
      key: "activeUsers",
      value: fmt(d.active_users),
      badge: null,
      trend: null,
    },
    {
      key: "mrrUsd",
      value: fmtUSD(d.mrr),
      badge: null,
      trend: null,
    },
    {
      key: "activeSubs",
      value: fmt(d.active_subs),
      badge: null,
      trend: null,
    },
    {
      key: "churnRisk",
      value: fmt(d.churn_risk),
      badge: "orange",
      trend: null,
    },
  ] as const;

  const lastUpdated = new Date(d.last_updated).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    timeZoneName: "short",
  });

  return (
    <div className="flex flex-col gap-6 p-4 md:p-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <p className="text-sm text-muted-foreground">
          {t("lastUpdated")} {lastUpdated}
        </p>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {kpiCards.map(({ key, value, badge }) => (
          <Card key={key} className="@container/card">
            <CardHeader className="pb-2">
              <CardDescription className="text-xs font-medium uppercase tracking-wide">
                {t(`kpi.${key}`)}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{value}</div>
            </CardContent>
            <CardFooter className="pt-0">
              {badge === "orange" ? (
                <Badge
                  variant="outline"
                  className="text-xs text-orange-600 border-orange-200 bg-orange-50"
                >
                  {t("kpi.dunningInProgress")}
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="gap-1 text-xs text-green-600 border-green-200 bg-green-50"
                >
                  <TrendingUp className="h-3 w-3" />
                  {t("kpi.vsLastMonth")}
                </Badge>
              )}
            </CardFooter>
          </Card>
        ))}
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <MrrTrendChart data={d.mrr_trend} activeSubs={d.active_subs} />
        </div>
        <SubStatusChart counts={d.status_counts} />
      </div>

      {/* Bottom row */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Recent Admin Actions */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t("recentActions.title")}</CardTitle>
            <CardDescription>admin_audit_log</CardDescription>
          </CardHeader>
          <CardContent className="p-0 px-2 pb-2">
            <AuditLogTable entries={d.audit_log} />
          </CardContent>
          <CardFooter>
            <a
              href="/dashboard/audit-log"
              className="text-xs text-primary hover:underline"
            >
              {t("recentActions.viewFullLog")} →
            </a>
          </CardFooter>
        </Card>

        {/* Webhook Health */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t("webhookHealth.title")}</CardTitle>
            <CardDescription>webhook_events</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {d.webhook_health.length === 0 ? (
              <p className="text-sm text-muted-foreground">No webhook events recorded.</p>
            ) : (
              d.webhook_health.map((p) => {
                const ok = p.Unprocessed === 0;
                return (
                  <div key={p.Provider} className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm font-medium capitalize">
                      {ok ? (
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                      ) : (
                        <AlertTriangle className="h-4 w-4 text-yellow-500" />
                      )}
                      {p.Provider}
                    </div>
                    {ok ? (
                      <Badge
                        variant="outline"
                        className="text-xs text-green-600 border-green-200 bg-green-50"
                      >
                        Healthy
                      </Badge>
                    ) : (
                      <Badge
                        variant="outline"
                        className="text-xs text-yellow-600 border-yellow-200 bg-yellow-50"
                      >
                        {p.Unprocessed} pending
                      </Badge>
                    )}
                  </div>
                );
              })
            )}
          </CardContent>
          <CardFooter>
            <p className="text-xs text-muted-foreground">
              {d.webhook_health.reduce((a, p) => a + p.Unprocessed, 0)}{" "}
              {t("webhookHealth.unprocessed")}
            </p>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
