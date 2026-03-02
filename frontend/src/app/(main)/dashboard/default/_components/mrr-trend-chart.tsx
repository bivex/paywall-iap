"use client";

import * as React from "react";
import { Area, AreaChart, CartesianGrid, XAxis } from "recharts";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { MonthlyMRR } from "@/actions/dashboard";

interface MrrTrendChartProps {
  data: MonthlyMRR[];
  activeSubs: number;
}

const chartConfig = {
  mrr: { label: "MRR (USD)", color: "var(--chart-1)" },
  subs: { label: "Active Subs", color: "var(--chart-2)" },
} satisfies ChartConfig;

// Build recharts-compatible data from API response
function buildChartData(trend: MonthlyMRR[], totalSubs: number) {
  return trend.map((m, i, arr) => ({
    date: m.Month,
    mrr: Math.round(m.MRR),
    // Distribute active subs linearly across months as approximation
    subs: Math.round((totalSubs / arr.length) * (i + 1)),
  }));
}

export function MrrTrendChart({ data, activeSubs }: MrrTrendChartProps) {
  const [metric, setMetric] = React.useState<"mrr" | "subs">("mrr");
  const chartData = buildChartData(data, activeSubs);

  return (
    <Card className="pt-0">
      <CardHeader className="flex items-center gap-2 space-y-0 border-b py-5 sm:flex-row">
        <div className="grid flex-1 gap-1">
          <CardTitle>Revenue &amp; Subscription Trend</CardTitle>
          <CardDescription>Last 6 months</CardDescription>
        </div>
        <Select value={metric} onValueChange={(v) => setMetric(v as "mrr" | "subs")}>
          <SelectTrigger
            className="hidden w-[160px] rounded-lg sm:ml-auto sm:flex"
            aria-label="Select metric"
          >
            <SelectValue placeholder="MRR (USD)" />
          </SelectTrigger>
          <SelectContent className="rounded-xl">
            <SelectItem value="mrr" className="rounded-lg">MRR (USD)</SelectItem>
            <SelectItem value="subs" className="rounded-lg">Active Subs</SelectItem>
          </SelectContent>
        </Select>
      </CardHeader>
      <CardContent className="px-2 pt-4 sm:px-6 sm:pt-6">
        <ChartContainer config={chartConfig} className="aspect-auto h-[200px] w-full">
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fillMrr" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--color-mrr)" stopOpacity={0.8} />
                <stop offset="95%" stopColor="var(--color-mrr)" stopOpacity={0.1} />
              </linearGradient>
              <linearGradient id="fillSubs" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--color-subs)" stopOpacity={0.8} />
                <stop offset="95%" stopColor="var(--color-subs)" stopOpacity={0.1} />
              </linearGradient>
            </defs>
            <CartesianGrid vertical={false} />
            <XAxis
              dataKey="date"
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tickFormatter={(v: string) => {
                const [year, month] = v.split("-");
                return new Date(Number(year), Number(month) - 1).toLocaleDateString("en-US", { month: "short" });
              }}
            />
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  labelFormatter={(v: string) => {
                    const [year, month] = v.split("-");
                    return new Date(Number(year), Number(month) - 1).toLocaleDateString("en-US", {
                      month: "long",
                      year: "numeric",
                    });
                  }}
                  formatter={(value, name) =>
                    name === "mrr"
                      ? [`$${Number(value).toLocaleString()}`, "MRR"]
                      : [Number(value).toLocaleString(), "Active Subs"]
                  }
                  indicator="dot"
                />
              }
            />
            <Area
              dataKey={metric}
              type="natural"
              fill={metric === "mrr" ? "url(#fillMrr)" : "url(#fillSubs)"}
              stroke={`var(--color-${metric})`}
              strokeWidth={2}
            />
            <ChartLegend content={<ChartLegendContent />} />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
