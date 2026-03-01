import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default async function AnalyticsPage() {
  const t = await getTranslations("analytics");
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <div className="flex gap-2">
          <Select><SelectTrigger className="w-48"><SelectValue placeholder={t("filter.dateRangePlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="30">{t("filter.last30")}</SelectItem><SelectItem value="90">{t("filter.last90")}</SelectItem><SelectItem value="365">{t("filter.last12m")}</SelectItem></SelectContent></Select>
          <Select><SelectTrigger className="w-40"><SelectValue placeholder={t("filter.dimensionPlaceholder")} /></SelectTrigger><SelectContent><SelectItem value="all">{t("filter.all")}</SelectItem><SelectItem value="platform">{t("filter.platform")}</SelectItem><SelectItem value="plan">{t("filter.plan")}</SelectItem></SelectContent></Select>
          <Button variant="outline" size="sm">{t("exportPng")}</Button>
          <Button variant="outline" size="sm">{t("exportPdf")}</Button>
        </div>
      </div>

      {/* KPI row */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {[
          { label: t("kpi.mrr"), value: "$12,450", delta: "▲ +8.3%", up: true },
          { label: t("kpi.churnRate"), value: "2.1%", delta: "▼ -0.4%", up: true },
          { label: t("kpi.ltv"), value: "$184.2", delta: "▲ +12.1%", up: true },
          { label: t("kpi.newSubs"), value: "412", delta: "▲ +5.7%", up: true },
        ].map((k) => (
          <Card key={k.label}>
            <CardHeader className="pb-2"><CardTitle className="text-xs font-medium text-muted-foreground uppercase">{k.label}</CardTitle></CardHeader>
            <CardContent><div className="text-2xl font-bold">{k.value}</div><p className="text-xs text-green-600 mt-1">{k.delta}</p></CardContent>
          </Card>
        ))}
      </div>

      {/* MRR Trend */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("mrrTrend.title")}</CardTitle></CardHeader>
        <CardContent>
          <div className="font-mono text-xs text-muted-foreground space-y-1">
            <div>$13k ─────────────────────────────────────────── ▲</div>
            <div>$12k ───────────────────────────────────── ▲</div>
            <div>$11k ──────────────────────────────── ▲</div>
            <div>$10k ─────────────────────── ▲</div>
            <div>  $9k ──────────── ▲</div>
            <div>  $8k ─── ▲</div>
            <div className="pt-1 flex justify-between"><span>Jan</span><span>Feb</span><span>Mar</span><span>Apr</span><span>May</span><span>Jun</span></div>
          </div>
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Churn by Platform */}
        <Card>
          <CardHeader><CardTitle className="text-sm">{t("churnByPlatform.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div>iOS &nbsp;&nbsp;&nbsp;&nbsp; <span className="font-mono">████████████ 1.8%</span></div>
            <div>Android <span className="font-mono">████████████████ 2.4%</span></div>
            <div>Web &nbsp;&nbsp;&nbsp;&nbsp; <span className="font-mono">██████████ 1.5%</span></div>
          </CardContent>
        </Card>
        {/* Revenue by Plan */}
        <Card>
          <CardHeader><CardTitle className="text-sm">{t("revenueByPlan.title")}</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="flex justify-between"><span>Pro Annual</span><span className="font-medium">$7,200 (58%)</span></div>
            <div className="flex justify-between"><span>Basic Monthly</span><span className="font-medium">$3,100 (25%)</span></div>
            <div className="flex justify-between"><span>Enterprise</span><span className="font-medium">$2,150 (17%)</span></div>
          </CardContent>
        </Card>
      </div>

      {/* Raw metrics table */}
      <Card>
        <CardHeader><CardTitle className="text-sm">{t("rawMetrics.title")}</CardTitle><p className="text-xs text-muted-foreground">analytics_snapshots</p></CardHeader>
        <CardContent>
          <Table>
            <TableHeader><TableRow><TableHead>{t("rawMetrics.period")}</TableHead><TableHead>{t("rawMetrics.metric")}</TableHead><TableHead className="text-right">{t("rawMetrics.value")}</TableHead><TableHead>{t("rawMetrics.delta")}</TableHead></TableRow></TableHeader>
            <TableBody>
              <TableRow><TableCell>2025-06</TableCell><TableCell>mrr</TableCell><TableCell className="text-right font-mono">$12,450</TableCell><TableCell className="text-green-600">▲ +8.3%</TableCell></TableRow>
              <TableRow><TableCell>2025-06</TableCell><TableCell>churn_rate</TableCell><TableCell className="text-right font-mono">0.021</TableCell><TableCell className="text-green-600">▼ -0.4%</TableCell></TableRow>
              <TableRow><TableCell>2025-06</TableCell><TableCell>ltv</TableCell><TableCell className="text-right font-mono">$184.20</TableCell><TableCell className="text-green-600">▲ +12.1%</TableCell></TableRow>
              <TableRow><TableCell>2025-06</TableCell><TableCell>new_subs</TableCell><TableCell className="text-right font-mono">412</TableCell><TableCell className="text-green-600">▲ +5.7%</TableCell></TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
