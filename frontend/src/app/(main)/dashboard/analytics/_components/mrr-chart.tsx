"use client";

import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import {
  TrendingDown, TrendingUp,
  DollarSign, BarChart2, Users, Percent,
} from "lucide-react";
import {
  Card, CardContent, CardDescription, CardHeader, CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig,
} from "@/components/ui/chart";
import { cn } from "@/lib/utils";
import type { TrendPoint } from "@/actions/analytics";

type MetricKey = "mrr" | "new_subs" | "active_count";

const chartConfig: ChartConfig = {
  mrr:          { label: "MRR (USD)",   color: "hsl(142 71% 45%)"  },
  new_subs:     { label: "New Subs",    color: "hsl(221 83% 53%)"  },
  active_count: { label: "Active Subs", color: "hsl(262 83% 58%)"  },
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
  if (pct == null) return null;
  return `${pct >= 0 ? "+" : ""}${pct.toFixed(1)}%`;
}

/* ─── Top KPI cards ──────────────────────────────────────────────── */
interface KpiCardProps {
  icon: React.ReactNode;
  label: string;
  value: string;
  delta?: string | null;
  up?: boolean;
  accent: string; // tailwind bg class for icon ring
  border: string; // tailwind border-l class
}

function KpiCard({ icon, label, value, delta, up, accent, border }: KpiCardProps) {
  return (
    <Card className={cn("relative overflow-hidden border-l-4 py-5", border)}>
      <CardContent className="px-5 py-0 flex items-start justify-between gap-3">
        <div className="flex flex-col gap-1">
          <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest">{label}</p>
          <p className="text-2xl font-bold tracking-tight">{value}</p>
          {delta != null && (
            <span className={cn("inline-flex items-center gap-1 text-xs font-medium mt-0.5", up ? "text-emerald-500" : "text-red-500")}>
              {up ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
              {delta} vs last 6 mo
            </span>
          )}
        </div>
        <div className={cn("flex h-10 w-10 shrink-0 items-center justify-center rounded-full", accent)}>
          {icon}
        </div>
      </CardContent>
    </Card>
  );
}

/* ─── Main export ────────────────────────────────────────────────── */
export function KpiAreaChart({ trend, mrr, ltv, churnRate, newSubsMonth }: Props) {
  const [active, setActive] = React.useState<MetricKey>("mrr");

  const topCards: KpiCardProps[] = [
    {
      label: "MRR",
      value: `$${mrr.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      delta: fmtDelta(pctDelta(trend, "mrr")),
      up: (pctDelta(trend, "mrr") ?? 0) >= 0,
      icon: <DollarSign className="h-5 w-5 text-emerald-600" />,
      accent: "bg-emerald-500/10",
      border: "border-l-emerald-500",
    },
    {
      label: "ARR",
      value: `$${(mrr * 12).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      icon: <BarChart2 className="h-5 w-5 text-blue-500" />,
      accent: "bg-blue-500/10",
      border: "border-l-blue-500",
    },
    {
      label: "Avg LTV",
      value: `$${ltv.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      icon: <Users className="h-5 w-5 text-violet-500" />,
      accent: "bg-violet-500/10",
      border: "border-l-violet-500",
    },
    {
      label: "Churn Rate",
      value: `${churnRate.toLocaleString("en-US")}%`,
      up: churnRate < 5,
      delta: churnRate === 0 ? null : churnRate < 5 ? "healthy" : "high risk",
      icon: <Percent className="h-5 w-5 text-amber-500" />,
      accent: churnRate >= 5 ? "bg-red-500/10" : "bg-amber-500/10",
      border: churnRate >= 5 ? "border-l-red-500" : "border-l-amber-500",
    },
  ];

  const metrics: { key: MetricKey; label: string; value: string; delta: string | null; up: boolean; color: string }[] = [
    {
      key:   "mrr",
      label: "MRR",
      value: `$${mrr.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      delta: fmtDelta(pctDelta(trend, "mrr")),
      up:    (pctDelta(trend, "mrr") ?? 0) >= 0,
      color: chartConfig.mrr.color as string,
    },
    {
      key:   "active_count",
      label: "Active Subs",
      value: String(trend[trend.length - 1]?.active_count ?? 0),
      delta: fmtDelta(pctDelta(trend, "active_count")),
      up:    (pctDelta(trend, "active_count") ?? 0) >= 0,
      color: chartConfig.active_count.color as string,
    },
    {
      key:   "new_subs",
      label: "New / Month",
      value: String(newSubsMonth),
      delta: fmtDelta(pctDelta(trend, "new_subs")),
      up:    (pctDelta(trend, "new_subs") ?? 0) >= 0,
      color: chartConfig.new_subs.color as string,
    },
  ];

  const m = metrics.find((x) => x.key === active)!;
  const fmt = (v: number) =>
    active === "mrr"
      ? `$${v.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`
      : String(v);

  return (
    <div className="space-y-4">
      {/* KPI summary row */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {topCards.map((k) => <KpiCard key={k.label} {...k} />)}
      </div>

      {/* Interactive chart card */}
      <Card className="overflow-hidden">
        {/* Metric tab selector */}
        <div className="grid grid-cols-3 divide-x border-b bg-muted/20">
          {metrics.map((metric) => {
            const isActive = active === metric.key;
            return (
              <button
                key={metric.key}
                onClick={() => setActive(metric.key)}
                className={cn(
                  "group relative flex flex-col gap-0.5 px-5 py-4 text-left transition-colors",
                  isActive ? "bg-background" : "hover:bg-muted/40",
                )}
              >
                <span className={cn("text-xs font-semibold uppercase tracking-widest transition-colors",
                  isActive ? "text-foreground" : "text-muted-foreground"
                )}>
                  {metric.label}
                </span>
                <span className={cn("text-xl font-bold tabular-nums",
                  isActive ? "text-foreground" : "text-muted-foreground/80"
                )}>
                  {metric.value}
                </span>
                {metric.delta && (
                  <span className={cn("flex items-center gap-1 text-xs font-medium",
                    metric.up ? "text-emerald-500" : "text-red-500"
                  )}>
                    {metric.up
                      ? <TrendingUp className="h-3 w-3" />
                      : <TrendingDown className="h-3 w-3" />}
                    {metric.delta} vs 6 mo
                  </span>
                )}
                {/* Active indicator bar */}
                <div
                  className={cn("absolute bottom-0 left-0 right-0 h-[2px] transition-opacity",
                    isActive ? "opacity-100" : "opacity-0"
                  )}
                  style={{ background: metric.color }}
                />
              </button>
            );
          })}
        </div>

        {/* Chart */}
        <CardContent className="px-3 pt-5 pb-3 sm:px-5">
          <div className="flex items-center gap-2 mb-4">
            <CardTitle className="text-sm font-semibold">{m.label} Trend</CardTitle>
            {m.delta && (
              <Badge
                variant="secondary"
                className={cn("text-xs px-2 py-0.5 font-medium border-0",
                  m.up
                    ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                    : "bg-red-500/10 text-red-600 dark:text-red-400"
                )}
              >
                {m.up ? <TrendingUp className="h-3 w-3 mr-1 inline" /> : <TrendingDown className="h-3 w-3 mr-1 inline" />}
                {m.delta}
              </Badge>
            )}
            <CardDescription className="ml-auto text-xs">Last 6 months · live data</CardDescription>
          </div>
          <ChartContainer config={chartConfig} className="aspect-auto h-[220px] w-full">
            <AreaChart data={trend} margin={{ left: 4, right: 4 }}>
              <defs>
                <linearGradient id={`grad-${active}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%"  stopColor={m.color} stopOpacity={0.25} />
                  <stop offset="95%" stopColor={m.color} stopOpacity={0.01} />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} strokeDasharray="3 3" className="stroke-border/50" />
              <XAxis
                dataKey="month"
                tickLine={false}
                axisLine={false}
                tickMargin={10}
                tickFormatter={(v: string) => v.slice(5)} /* show MM only */
                className="text-xs"
              />
              <YAxis hide />
              <ChartTooltip
                content={<ChartTooltipContent formatter={(v) => [fmt(Number(v)), m.label]} />}
              />
              <Area
                dataKey={active}
                type="monotone"
                stroke={m.color}
                strokeWidth={2.5}
                fill={`url(#grad-${active})`}
                dot={{ r: 4, fill: m.color, strokeWidth: 0 }}
                activeDot={{ r: 6, strokeWidth: 2, stroke: "var(--background)" }}
              />
            </AreaChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  );
}
