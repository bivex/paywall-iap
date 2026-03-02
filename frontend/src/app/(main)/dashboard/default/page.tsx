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
import { Separator } from "@/components/ui/separator";

import { MrrTrendChart } from "./_components/mrr-trend-chart";
import { SubStatusChart } from "./_components/sub-status-chart";

const kpiData = [
  {
    key: "activeUsers",
    value: "14,205",
    trend: "+2.1%",
    positive: true,
    trendLabel: "vsLastMonth",
  },
  {
    key: "mrrUsd",
    value: "$45,230",
    trend: "+8.3%",
    positive: true,
    trendLabel: "vsLastMonth",
  },
  {
    key: "activeSubs",
    value: "12,100",
    trend: "+5.2%",
    positive: true,
    trendLabel: "vsLastMonth",
  },
  {
    key: "churnRisk",
    value: "345",
    trend: null,
    positive: false,
    trendLabel: "dunningInProgress",
  },
];

const auditLog = [
  { time: "13:40", actor: "Admin_01", action: "updated pricing tier", detail: "Pro Annual → $39.99" },
  { time: "13:35", actor: "Admin_02", action: "refunded transaction", detail: "txn_8821 · $49.99" },
  { time: "13:30", actor: "System", action: "auto-retry dunning", detail: "sub_3341 · attempt 2/4" },
];

const webhookProviders = [
  { name: "Stripe", ok: true },
  { name: "Apple", ok: true },
  { name: "Google", ok: false, detail: "2 failed, retrying" },
];

export default async function DashboardPage() {
  const t = await getTranslations("dashboard");

  return (
    <div className="flex flex-col gap-6 p-4 md:p-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <p className="text-sm text-muted-foreground">
          {t("lastUpdated")} 2026-03-01 13:40 UTC
        </p>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {kpiData.map((kpi) => (
          <Card key={kpi.key} className="@container/card">
            <CardHeader className="pb-2">
              <CardDescription className="text-xs font-medium uppercase tracking-wide">
                {t(`kpi.${kpi.key}`)}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{kpi.value}</div>
            </CardContent>
            <CardFooter className="pt-0">
              {kpi.trend ? (
                <Badge
                  variant="outline"
                  className="gap-1 text-xs text-green-600 border-green-200 bg-green-50"
                >
                  <TrendingUp className="h-3 w-3" />
                  {kpi.trend} {t("kpi.vsLastMonth")}
                </Badge>
              ) : (
                <Badge
                  variant="outline"
                  className="text-xs text-orange-600 border-orange-200 bg-orange-50"
                >
                  {t("kpi.dunningInProgress")}
                </Badge>
              )}
            </CardFooter>
          </Card>
        ))}
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <MrrTrendChart />
        </div>
        <SubStatusChart />
      </div>

      {/* Bottom row */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Recent Admin Actions */}
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t("recentActions.title")}</CardTitle>
            <CardDescription>admin_audit_log</CardDescription>
          </CardHeader>
          <CardContent className="space-y-0">
            {auditLog.map((entry, i) => (
              <div key={i}>
                <div className="flex items-start gap-2 py-2 text-sm">
                  <span className="text-muted-foreground tabular-nums shrink-0">
                    [{entry.time}]
                  </span>
                  <div>
                    <span className="font-medium">{entry.actor}</span>{" "}
                    <span className="text-muted-foreground">{entry.action}</span>{" "}
                    <span className="font-medium">{entry.detail}</span>
                  </div>
                </div>
                {i < auditLog.length - 1 && <Separator />}
              </div>
            ))}
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
            {webhookProviders.map((p) => (
              <div key={p.name} className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-sm font-medium">
                  {p.ok ? (
                    <CheckCircle2 className="h-4 w-4 text-green-500" />
                  ) : (
                    <AlertTriangle className="h-4 w-4 text-yellow-500" />
                  )}
                  {p.name}
                </div>
                {p.ok ? (
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
                    {p.detail}
                  </Badge>
                )}
              </div>
            ))}
          </CardContent>
          <CardFooter>
            <p className="text-xs text-muted-foreground">
              0 {t("webhookHealth.unprocessed")}
            </p>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
