"use client";

import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { TrendingDown, TrendingUp } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import { cn } from "@/lib/utils";
import type { TrendPoint } from "@/actions/analytics";

type MetricKey = "mrr" | "new_subs" | "active_count";

interface KpiDef {
  key: MetricKey;
  label: string;
  value: string;
  delta: string;
  up: boolean;
  format: (v: number) => string;
  color: string;
}

const chartConfig: ChartConfig = {
  mrr:          { label: "MRR (USD)",   color: "var(--chart-1)" },
  new_subs:     { label: "New Subs",    color: "var(--chart-2)" },
  active_count: { label: "Active Subs", color: "var(--chart-3)" },
};

interface Props {
  trend: TrendPoint[];
  mrr: number;
  ltv: number;
  churnRate: number;
  newSubsMonth: number;
}

function pctDelta(trend: TrendPoint[], key: MetricKey): number | null {
  const first = trend[0];
  const last  = trend[trend.length - 1];
  if (!first || !last || first[key] === 0) return null;
  return ((last[key] - first[key]) / first[key]) * 100;
}

function fmtDelta(pct: number | null) {
  if (pct == null) return "—";
  return `${pct >= 0 ? "+" : ""}${pct.toFixed(1)}%`;
}

export function KpiAreaChart({ trend, mrr, ltv, churnRate, newSubsMonth }: Props) {
  const [active, setActive] = React.useState<MetricKey>("mrr");

  const kpis: KpiDef[] = [
    {
      key: "mrr",
      label: "MRR",
      value: `$${mrr.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      delta: fmtDelta(pctDelta(trend, "mrr")),
      up: (pctDelta(trend, "mrr") ?? 0) >= 0,
      format: (v) => `$${v.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      color: "var(--chart-1)",
    },
    {
      key: "active_count",
      label: "Active Subs",
      value: String(trend[trend.length - 1]?.active_count ?? 0),
      delta: fmtDelta(pctDelta(trend, "active_count")),
      up: (pctDelta(trend, "active_count") ?? 0) >= 0,
      format: (v) => String(v),
      color: "var(--chart-3)",
    },
    {
      key: "new_subs",
      label: "New / Month",
      value: String(newSubsMonth),
      delta: fmtDelta(pctDelta(trend, "new_subs")),
      up: (pctDelta(trend, "new_subs") ?? 0) >= 0,
      format: (v) => String(v),
      color: "var(--chart-2)",
    },
  ];

  const m = kpis.find((x) => x.key === active)!;

  return (
    <div className="space-y-4">
      {/* KPI summary row */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          { label: "MRR",        value: `$${mrr.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`, muted: false },
          { label: "ARR",        value: `$${(mrr * 12).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`, muted: false },
          { label: "Avg LTV",    value: `$${ltv.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`, muted: false },
          { label: "Churn Rate", value: `${churnRate.toLocaleString("en-US")}%`, muted: churnRate >= 5 },
        ].map((k) => (
          <Card key={k.label} className="py-4">
            <CardContent className="px-4 py-0">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1">{k.label}</p>
              <p className={cn("text-xl font-bold", k.muted && "text-red-500")}>{k.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Chart with metric selector */}
      <Card>
        <CardHeader className="p-0 border-b">
          <div className="flex divide-x">
            {kpis.map((metric) => (
              <button
                key={metric.key}
                onClick={() => setActive(metric.key)}
                data-active={active === metric.key}
                className={cn(
                  "relative flex-1 flex flex-col gap-0.5 px-5 py-4 text-left transition-colors",
                  "hover:bg-muted/30 data-[active=true]:bg-muted/50"
                )}
              >
                <span className="text-xs text-muted-foreground font-medium">{metric.label}</span>
                <span className="text-lg font-bold">{metric.value}</span>
                <span className={cn("text-xs flex items-center gap-1", metric.up ? "text-emerald-600" : "text-red-500")}>
                  {metric.up ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
                  {metric.delta} vs 6 mo
                </span>
                {active === metric.key && (
                  <div className="absolute bottom-0 left-0 right-0 h-0.5" style={{ background: metric.color }} />
                )}
              </button>
            ))}
          </div>
        </CardHeader>
        <CardContent className="pt-4 px-2 sm:px-4">
          <div className="flex items-center gap-2 mb-3 px-2">
            <CardTitle className="text-sm">{m.label} Trend</CardTitle>
            <Badge variant="outline" className={cn(
              "text-xs border-none",
              m.up ? "text-emerald-600 bg-emerald-50 dark:bg-emerald-950/30" : "text-red-500 bg-red-50"
            )}>
              {m.up ? <TrendingUp className="h-3 w-3 mr-1 inline" /> : <TrendingDown className="h-3 w-3 mr-1 inline" />}
              {m.delta}
            </Badge>
            <CardDescription className="ml-auto text-xs">Real DB data</CardDescription>
          </div>
          <ChartContainer config={chartConfig} className="aspect-auto h-[240px] w-full">
            <AreaChart data={trend} margin={{ left: 8, right: 8 }}>
              <defs>
                <linearGradient id={`grad-${active}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%"  stopColor={m.color} stopOpacity={0.3} />
                  <stop offset="95%" stopColor={m.color} stopOpacity={0.02} />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} strokeDasharray="3 3" />
              <XAxis dataKey="month" tickLine={false} axisLine={false} tickMargin={8} />
              <YAxis hide />
              <ChartTooltip content={<ChartTooltipContent formatter={(v) => [m.format(Number(v)), m.label]} />} />
              <Area
                dataKey={active}
                type="monotone"
                stroke={m.color}
                strokeWidth={2}
                fill={`url(#grad-${active})`}
                dot={{ r: 3.5, fill: m.color, strokeWidth: 0 }}
                activeDot={{ r: 5 }}
              />
            </AreaChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  );
}
