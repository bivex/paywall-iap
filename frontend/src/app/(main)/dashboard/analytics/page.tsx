import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { TrendingUp, TrendingDown, AlertCircle, CheckCircle2, Clock, XCircle, AlertTriangle } from "lucide-react";
import { getAnalyticsReport } from "@/actions/analytics";
import { KpiAreaChart } from "./_components/mrr-chart";
import { PlatformBarChart } from "./_components/platform-bar-chart";
import { RevenueDonutChart } from "./_components/revenue-donut-chart";

export default async function AnalyticsPage() {
  const report = await getAnalyticsReport();

  if (!report) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-24 text-muted-foreground">
        <AlertCircle className="h-8 w-8" />
        <p className="text-sm">Failed to load analytics — make sure you are logged in.</p>
      </div>
    );
  }

  const { mrr, arr, ltv, total_revenue, churn_rate, new_subs_month, trend, by_platform, by_plan, status_counts } = report;

  const statusRows = [
    { label: "Active",    value: status_counts.active,    icon: CheckCircle2, color: "text-emerald-500", bg: "bg-emerald-500/10" },
    { label: "Grace",     value: status_counts.grace,     icon: Clock,        color: "text-amber-500",   bg: "bg-amber-500/10"   },
    { label: "Cancelled", value: status_counts.cancelled, icon: XCircle,      color: "text-slate-400",   bg: "bg-slate-500/10"   },
    { label: "Expired",   value: status_counts.expired,   icon: AlertTriangle,color: "text-red-500",     bg: "bg-red-500/10"     },
  ];

  const metricsTable = [
    { metric: "mrr",            formula: "Σ(active subs × monthly price)",       value: `$${mrr.toLocaleString("en-US", { minimumFractionDigits: 2 })}`,           up: true             },
    { metric: "arr",            formula: "MRR × 12",                              value: `$${arr.toLocaleString("en-US", { minimumFractionDigits: 2 })}`,           up: true             },
    { metric: "ltv",            formula: "total_revenue / distinct paying users", value: `$${ltv.toLocaleString("en-US", { minimumFractionDigits: 2 })}`,           up: true             },
    { metric: "total_revenue",  formula: "Σ successful transactions",             value: `$${total_revenue.toLocaleString("en-US", { minimumFractionDigits: 2 })}`, up: true             },
    { metric: "churn_rate",     formula: "churned_mo / (active + churned) × 100",value: `${churn_rate}%`,                                                           up: churn_rate < 5   },
    { metric: "new_subs_month", formula: "COUNT new subs created this month",     value: String(new_subs_month),                                                     up: new_subs_month > 0 },
  ];

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Analytics Reports</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            Live DB data · formulas: MRR = subs × price · LTV = revenue / users · Churn = churned / total
          </p>
        </div>
        <Badge variant="outline" className="text-xs text-emerald-600 border-emerald-500/40 bg-emerald-500/5 shrink-0">
          <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 mr-1.5 inline-block" />
          Live
        </Badge>
      </div>

      {/* KPI cards + trend chart */}
      <KpiAreaChart
        trend={trend ?? []}
        mrr={mrr}
        ltv={ltv}
        churnRate={churn_rate}
        newSubsMonth={new_subs_month}
      />

      {/* Subscription status row */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {statusRows.map((s) => {
          const Icon = s.icon;
          return (
            <Card key={s.label} className="py-4">
              <CardContent className="px-4 py-0 flex items-center justify-between">
                <div>
                  <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">{s.label}</p>
                  <p className={`text-2xl font-bold ${s.color}`}>{s.value}</p>
                </div>
                <div className={`flex h-9 w-9 items-center justify-center rounded-full ${s.bg}`}>
                  <Icon className={`h-4 w-4 ${s.color}`} />
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Platform bar + Plan donut */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <PlatformBarChart data={by_platform ?? []} />
        </div>
        <RevenueDonutChart data={by_plan ?? []} totalMrr={mrr} />
      </div>

      {/* Metrics formula table */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center gap-2">
            <CardTitle className="text-sm font-semibold">Metrics & Formulas</CardTitle>
            <Badge variant="secondary" className="text-xs">computed from DB</Badge>
          </div>
        </CardHeader>
        <CardContent className="pt-0">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-36">Metric</TableHead>
                <TableHead>Formula</TableHead>
                <TableHead className="text-right w-36">Value</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {metricsTable.map((row) => (
                <TableRow key={row.metric}>
                  <TableCell>
                    <Badge variant="secondary" className="font-mono text-xs">{row.metric}</Badge>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground font-mono">{row.formula}</TableCell>
                  <TableCell className="text-right">
                    <span className={`inline-flex items-center gap-1 text-sm font-semibold tabular-nums ${row.up ? "text-emerald-500" : "text-red-500"}`}>
                      {row.up ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                      {row.value}
                    </span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
