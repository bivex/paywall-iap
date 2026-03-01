import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default async function DashboardPage() {
  const t = await getTranslations("dashboard");
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <p className="text-sm text-muted-foreground">{t("lastUpdated")} 2026-03-01 13:40 UTC</p>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t("kpi.activeUsers")}</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">14,205</div><p className="text-xs text-green-600 mt-1">▲ +2.1% {t("kpi.vsLastMonth")}</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t("kpi.mrrUsd")}</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">$45,230</div><p className="text-xs text-green-600 mt-1">▲ +8.3% {t("kpi.vsLastMonth")}</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t("kpi.activeSubs")}</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">12,100</div><p className="text-xs text-green-600 mt-1">▲ +5.2% {t("kpi.vsLastMonth")}</p></CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{t("kpi.churnRisk")}</CardTitle></CardHeader>
          <CardContent><div className="text-2xl font-bold">345</div><p className="text-xs text-orange-600 mt-1">{t("kpi.dunningInProgress")}</p></CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Recent Admin Actions */}
        <Card>
          <CardHeader><CardTitle className="text-sm">{t("recentActions.title")}</CardTitle><p className="text-xs text-muted-foreground">admin_audit_log</p></CardHeader>
          <CardContent className="space-y-2">
            <div className="text-sm">[13:40] <span className="font-medium">Admin_01</span> updated pricing tier <span className="text-muted-foreground">Pro Annual → $39.99</span></div>
            <Separator />
            <div className="text-sm">[13:35] <span className="font-medium">Admin_02</span> refunded transaction <span className="text-muted-foreground">txn_8821 · $49.99</span></div>
            <Separator />
            <div className="text-sm">[13:30] <span className="font-medium">System</span> auto-retry dunning <span className="text-muted-foreground">sub_3341 · attempt 2/4</span></div>
            <a href="/dashboard/audit-log" className="text-xs text-primary mt-2 block">{t("recentActions.viewFullLog")}</a>
          </CardContent>
        </Card>

        {/* Subscription Status Distribution */}
        <Card>
          <CardHeader><CardTitle className="text-sm">{t("subStatusDist.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            {[
              { label: t("subStatusDist.active"), value: 85, color: "bg-green-500" },
              { label: t("subStatusDist.gracePeriod"), value: 5, color: "bg-yellow-500" },
              { label: t("subStatusDist.cancelled"), value: 7, color: "bg-gray-400" },
              { label: t("subStatusDist.expired"), value: 3, color: "bg-red-500" },
            ].map((item) => (
              <div key={item.label} className="space-y-1">
                <div className="flex justify-between text-xs"><span>{item.label}</span><span className="text-muted-foreground">{item.value}%</span></div>
                <Progress value={item.value} className="h-2" />
              </div>
            ))}
          </CardContent>
        </Card>
      </div>

      {/* Webhook Health */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("webhookHealth.title")}</CardTitle><p className="text-xs text-muted-foreground">webhook_events</p></CardHeader>
        <CardContent>
          <div className="flex gap-4 text-sm mb-2">
            <span>✅ Stripe</span><span>✅ Apple</span><span>⚠️ {t("webhookHealth.googleFailed")}</span>
          </div>
          <p className="text-xs text-muted-foreground">{t("webhookHealth.unprocessed")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
