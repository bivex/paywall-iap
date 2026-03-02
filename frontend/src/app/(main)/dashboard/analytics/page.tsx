import { TrendingDown, TrendingUp } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { KpiAreaChart } from "./_components/mrr-chart";
import { PlatformBarChart } from "./_components/platform-bar-chart";
import { RevenueDonutChart } from "./_components/revenue-donut-chart";
import { ChurnBarChart } from "./_components/churn-bar-chart";

const rawMetrics = [
  { period: "2026-01", metric: "mrr",        value: "$12,450", delta: "+8.3%",  up: true  },
  { period: "2026-01", metric: "churn_rate", value: "1.8%",    delta: "-0.6pp", up: true  },
  { period: "2026-01", metric: "ltv",        value: "$184",    delta: "+31.4%", up: true  },
  { period: "2026-01", metric: "new_subs",   value: "412",     delta: "+32%",   up: true  },
  { period: "2025-12", metric: "mrr",        value: "$11,430", delta: "+5.8%",  up: true  },
  { period: "2025-12", metric: "churn_rate", value: "1.9%",    delta: "-0.2pp", up: true  },
];

export default async function AnalyticsPage() {
  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-semibold">Analytics Reports</h1>
        <p className="text-sm text-muted-foreground">analytics_snapshots · last 6 months · click a metric to explore</p>
      </div>

      {/* KPI selector + Area chart (combined) */}
      <KpiAreaChart />

      {/* Platform bar + Revenue donut */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <PlatformBarChart />
        </div>
        <RevenueDonutChart />
      </div>

      {/* Churn by platform */}
      <ChurnBarChart />

      {/* Raw metrics table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm">Raw Metrics Snapshot</CardTitle>
          <p className="text-xs text-muted-foreground">analytics_snapshots</p>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Period</TableHead>
                <TableHead>Metric</TableHead>
                <TableHead className="text-right">Value</TableHead>
                <TableHead>Delta</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rawMetrics.map((row, i) => (
                <TableRow key={i}>
                  <TableCell className="font-mono text-xs">{row.period}</TableCell>
                  <TableCell>
                    <Badge variant="secondary" className="font-mono text-xs">{row.metric}</Badge>
                  </TableCell>
                  <TableCell className="text-right font-mono">{row.value}</TableCell>
                  <TableCell>
                    <span className={`text-xs flex items-center gap-1 ${row.up ? "text-emerald-600" : "text-red-500"}`}>
                      {row.up ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                      {row.delta}
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
