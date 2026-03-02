"use client";

import * as React from "react";
import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts";
import {
  Card, CardContent, CardDescription, CardHeader, CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig,
} from "@/components/ui/chart";

const data = [
  { month: "Aug", mrr: 8100, subs: 312 },
  { month: "Sep", mrr: 8920, subs: 345 },
  { month: "Oct", mrr: 9750, subs: 368 },
  { month: "Nov", mrr: 10800, subs: 390 },
  { month: "Dec", mrr: 11430, subs: 402 },
  { month: "Jan", mrr: 12450, subs: 412 },
];

const chartConfig = {
  mrr: { label: "MRR (USD)", color: "var(--chart-1)" },
  subs: { label: "Active Subs", color: "var(--chart-2)" },
} satisfies ChartConfig;

type Key = "mrr" | "subs";

export function MrrChart() {
  const [active, setActive] = React.useState<Key>("mrr");

  const total = React.useMemo(
    () => ({
      mrr: data[data.length - 1].mrr,
      subs: data[data.length - 1].subs,
    }),
    []
  );

  return (
    <Card className="py-0">
      <CardHeader className="flex flex-col items-stretch border-b !p-0 sm:flex-row">
        <div className="flex flex-1 flex-col justify-center gap-1 px-6 pt-4 pb-3 sm:!py-0">
          <CardTitle>Revenue & Subscription Trend</CardTitle>
          <CardDescription>Last 6 months</CardDescription>
        </div>
        <div className="flex">
          {(["mrr", "subs"] as Key[]).map((key) => (
            <button
              key={key}
              data-active={active === key}
              className="data-[active=true]:bg-muted/50 relative flex flex-1 flex-col justify-center gap-1 border-t px-6 py-4 text-left even:border-l sm:border-t-0 sm:border-l sm:px-8 sm:py-6"
              onClick={() => setActive(key)}
            >
              <span className="text-muted-foreground text-xs">{chartConfig[key].label}</span>
              <span className="text-lg font-bold sm:text-3xl">
                {key === "mrr" ? `$${total.mrr.toLocaleString()}` : total.subs.toLocaleString()}
              </span>
            </button>
          ))}
        </div>
      </CardHeader>
      <CardContent className="px-2 sm:p-6">
        <ChartContainer config={chartConfig} className="aspect-auto h-[250px] w-full">
          <LineChart data={data} margin={{ left: 12, right: 12 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="month" tickLine={false} axisLine={false} tickMargin={8} />
            <YAxis hide />
            <ChartTooltip
              content={
                <ChartTooltipContent
                  className="w-[160px]"
                  formatter={(value, name) =>
                    name === "mrr"
                      ? [`$${Number(value).toLocaleString()}`, "MRR"]
                      : [value, "Active Subs"]
                  }
                />
              }
            />
            <Line
              dataKey={active}
              type="monotone"
              stroke={`var(--color-${active})`}
              strokeWidth={2}
              dot={{ r: 4, fill: `var(--color-${active})` }}
            />
          </LineChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
