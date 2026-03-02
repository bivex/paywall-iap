"use client";

import * as React from "react";
import { Bar, BarChart, CartesianGrid, XAxis } from "recharts";
import {
  Card, CardContent, CardDescription, CardHeader, CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig,
} from "@/components/ui/chart";

const data = [
  { month: "Aug", ios: 120, android: 95, web: 97 },
  { month: "Sep", ios: 134, android: 110, web: 101 },
  { month: "Oct", ios: 145, android: 118, web: 105 },
  { month: "Nov", ios: 152, android: 130, web: 108 },
  { month: "Dec", ios: 158, android: 138, web: 106 },
  { month: "Jan", ios: 162, android: 145, web: 105 },
];

const chartConfig = {
  ios: { label: "iOS", color: "var(--chart-1)" },
  android: { label: "Android", color: "var(--chart-2)" },
  web: { label: "Web", color: "var(--chart-3)" },
} satisfies ChartConfig;

type Key = "ios" | "android" | "web";

export function PlatformBarChart() {
  const [active, setActive] = React.useState<Key>("ios");

  const total = React.useMemo(
    () => ({
      ios: data.reduce((s, d) => s + d.ios, 0),
      android: data.reduce((s, d) => s + d.android, 0),
      web: data.reduce((s, d) => s + d.web, 0),
    }),
    []
  );

  return (
    <Card className="py-0">
      <CardHeader className="flex flex-col items-stretch border-b !p-0 sm:flex-row">
        <div className="flex flex-1 flex-col justify-center gap-1 px-6 pt-4 pb-3 sm:!py-0">
          <CardTitle>New Subscriptions by Platform</CardTitle>
          <CardDescription>Last 6 months — click to switch</CardDescription>
        </div>
        <div className="flex">
          {(["ios", "android", "web"] as Key[]).map((key) => (
            <button
              key={key}
              data-active={active === key}
              className="data-[active=true]:bg-muted/50 flex flex-1 flex-col justify-center gap-1 border-t px-4 py-3 text-left even:border-l sm:border-t-0 sm:border-l sm:px-6 sm:py-4"
              onClick={() => setActive(key)}
            >
              <span className="text-muted-foreground text-xs">{chartConfig[key].label}</span>
              <span className="text-base font-bold sm:text-2xl">{total[key]}</span>
            </button>
          ))}
        </div>
      </CardHeader>
      <CardContent className="px-2 sm:p-6">
        <ChartContainer config={chartConfig} className="aspect-auto h-[250px] w-full">
          <BarChart data={data} margin={{ left: 12, right: 12 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="month" tickLine={false} axisLine={false} tickMargin={8} />
            <ChartTooltip content={<ChartTooltipContent />} />
            <Bar dataKey={active} fill={`var(--color-${active})`} radius={4} />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
