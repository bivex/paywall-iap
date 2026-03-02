"use client";

import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { TrendingDown, TrendingUp } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import { cn } from "@/lib/utils";

const data = [
  { month: "Aug", mrr: 8100,  churn: 2.4, ltv: 140, new_subs: 312 },
  { month: "Sep", mrr: 8920,  churn: 2.2, ltv: 152, new_subs: 345 },
  { month: "Oct", mrr: 9750,  churn: 2.1, ltv: 161, new_subs: 368 },
  { month: "Nov", mrr: 10800, churn: 2.0, ltv: 170, new_subs: 390 },
  { month: "Dec", mrr: 11430, churn: 1.9, ltv: 178, new_subs: 402 },
  { month: "Jan", mrr: 12450, churn: 1.8, ltv: 184, new_subs: 412 },
];

type MetricKey = "mrr" | "churn" | "ltv" | "new_subs";

const metrics: {
  key: MetricKey;
  label: string;
  value: string;
  delta: string;
  up: boolean;
  format: (v: number) => string;
  color: string;
}[] = [
  { key: "mrr",      label: "MRR",        value: "$12,450", delta: "+8.3%",  up: true,  format: (v) => `$${v.toLocaleString("en-US")}`, color: "var(--chart-1)" },
  { key: "churn",    label: "Churn Rate", value: "1.8%",    delta: "-0.6pp", up: true,  format: (v) => `${v}%`,                          color: "var(--chart-2)" },
  { key: "ltv",      label: "Avg LTV",    value: "$184",    delta: "+31.4%", up: true,  format: (v) => `$${v}`,                           color: "var(--chart-3)" },
  { key: "new_subs", label: "New Subs",   value: "412",     delta: "+32%",   up: true,  format: (v) => String(v),                         color: "var(--chart-4)" },
];

const chartConfig: ChartConfig = {
  mrr:      { label: "MRR (USD)",    color: "var(--chart-1)" },
  churn:    { label: "Churn Rate %", color: "var(--chart-2)" },
  ltv:      { label: "Avg LTV",      color: "var(--chart-3)" },
  new_subs: { label: "New Subs",     color: "var(--chart-4)" },
};

export function KpiAreaChart() {
  const [active, setActive] = React.useState<MetricKey>("mrr");
  const m = metrics.find((x) => x.key === active)!;

  return (
    <Card>
      {/* KPI selector row */}
      <CardHeader className="p-0 border-b">
        <div className="grid grid-cols-2 lg:grid-cols-4">
          {metrics.map((metric) => (
            <button
              key={metric.key}
              onClick={() => setActive(metric.key)}
              data-active={active === metric.key}
              className={cn(
                "flex flex-col gap-1 p-5 text-left border-r last:border-r-0 transition-colors",
                "hover:bg-muted/40",
                "data-[active=true]:bg-muted/60"
              )}
            >
              <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                {metric.label}
              </span>
              <span className="text-2xl font-bold">{metric.value}</span>
              <span className={cn(
                "text-xs flex items-center gap-1",
                metric.up ? "text-emerald-600" : "text-red-500"
              )}>
                {metric.up
                  ? <TrendingUp className="h-3 w-3" />
                  : <TrendingDown className="h-3 w-3" />}
                {metric.delta} vs 6 mo ago
              </span>
              {active === metric.key && (
                <div
                  className="absolute bottom-0 left-0 right-0 h-0.5 rounded-full"
                  style={{ background: metric.color }}
                />
              )}
            </button>
          ))}
        </div>
      </CardHeader>

      <CardContent className="pt-6 px-2 sm:px-6">
        <div className="mb-3 flex items-center gap-2">
          <CardTitle className="text-sm">{m.label} Trend</CardTitle>
          <Badge
            variant="outline"
            className={cn(
              "text-xs",
              m.up ? "text-emerald-600 border-emerald-200 bg-emerald-50 dark:bg-emerald-950/30" : "text-red-500"
            )}
          >
            {m.up ? <TrendingUp className="h-3 w-3 mr-1" /> : <TrendingDown className="h-3 w-3 mr-1" />}
            {m.delta}
          </Badge>
          <CardDescription className="ml-auto text-xs">Aug – Jan 2026</CardDescription>
        </div>

        <ChartContainer config={chartConfig} className="aspect-auto h-[260px] w-full">
          <AreaChart data={data} margin={{ left: 12, right: 12 }}>
            <defs>
              <linearGradient id={`grad-${active}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%"  stopColor={m.color} stopOpacity={0.35} />
                <stop offset="95%" stopColor={m.color} stopOpacity={0.02} />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} strokeDasharray="3 3" />
            <XAxis dataKey="month" tickLine={false} axisLine={false} tickMargin={8} />
            <YAxis hide />
            <ChartTooltip
              content={
                <ChartTooltipContent
                  formatter={(value) => [m.format(Number(value)), m.label]}
                />
              }
            />
            <Area
              dataKey={active}
              type="monotone"
              stroke={m.color}
              strokeWidth={2}
              fill={`url(#grad-${active})`}
              dot={{ r: 4, fill: m.color, strokeWidth: 0 }}
              activeDot={{ r: 6 }}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
