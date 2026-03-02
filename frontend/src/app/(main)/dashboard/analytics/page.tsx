import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { TrendingUp, TrendingDown, AlertCircle } from "lucide-react";
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

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-semibold">Analytics Reports</h1>
        <p className="text-sm text-muted-foreground">
          Real DB data · MRR = active subs × price/mo · LTV = revenue / users · Churn = churned / total
        </p>
      </div>

      {/* KPI cards + area chart (metric selector) */}
      <KpiAreaChart
        trend={trend ?? []}
        mrr={mrr}
        ltv={ltv}
        churnRate={churn_rate}
        newSubsMonth={new_subs_month}
      />

      {/* Platform bars + Plan donut */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <PlatformBarChart data={by_platform ?? []} />
        </div>
        <RevenueDonutChart data={by_plan ?? []} totalMrr={mrr} />
      </div>

      {/* Sub status + revenue summary */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {[
          { label: "Active",    value: status_counts.active,    color: "text-emerald-600" },
          { label: "Grace",     value: status_counts.grace,     color: "text-yellow-600"  },
          { label: "Cancelled", value: status_counts.cancelled, color: "text-muted-foreground" },
          { label: "Expired",   value: status_counts.expired,   color: "text-red-500"     },
        ].map((s) => (
          <Card key={s.label} className="py-4">
            <CardContent className="px-4 py-0">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">{s.label}</p>
              <p className={`text-2xl font-bold ${s.color}`}>{s.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Raw metrics summary table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Metrics Summary</CardTitle>
          <p className="text-xs text-muted-foreground">Computed from subscriptions + transactions · formulas in tooltip</p>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Metric</TableHead>
                <TableHead>Formula</TableHead>
                <TableHead className="text-right">Value</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {[
                { metric: "mrr",           formula: "Σ(active subs × monthly price)",          value: `$${mrr.toLocaleString("en-US", { minimumFractionDigits: 2 })}`,           up: true  },
                { metric: "arr",           formula: "MRR × 12",                                 value: `$${arr.toLocaleString("en-US", { minimumFractionDigits: 2 })}`,           up: true  },
                { metric: "ltv",           formula: "total_revenue / distinct users",            value: `$${ltv.toLocaleString("en-US", { minimumFractionDigits: 2 })}`,           up: true  },
                { metric: "total_revenue", formula: "Σ successful transactions",                 value: `$${total_revenue.toLocaleString("en-US", { minimumFractionDigits: 2 })}`, up: true  },
                { metric: "churn_rate",    formula: "churned_mo / (active + churned) × 100",    value: `${churn_rate}%`,                                                           up: churn_rate < 5 },
                { metric: "new_subs_month",formula: "COUNT new subs created this month",        value: String(new_subs_month),                                                     up: true  },
              ].map((row) => (
                <TableRow key={row.metric}>
                  <TableCell><Badge variant="secondary" className="font-mono text-xs">{row.metric}</Badge></TableCell>
                  <TableCell className="text-xs text-muted-foreground font-mono">{row.formula}</TableCell>
                  <TableCell className="text-right font-mono text-sm">
                    <span className={`flex items-center justify-end gap-1 ${row.up ? "text-emerald-600" : "text-red-500"}`}>
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
