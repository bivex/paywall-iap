"use client";

import { Label, Pie, PieChart, Sector } from "recharts";
import type { PieSectorDataItem } from "recharts/types/polar/Pie";

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "@/components/ui/chart";
import type { SubscriptionStatusCounts } from "@/actions/dashboard";

interface SubStatusChartProps {
  counts: SubscriptionStatusCounts;
}

const chartConfig = {
  count: { label: "Subscriptions" },
  active: { label: "Active", color: "var(--chart-1)" },
  grace: { label: "Grace Period", color: "var(--chart-2)" },
  cancelled: { label: "Cancelled", color: "var(--chart-3)" },
  expired: { label: "Expired", color: "var(--chart-4)" },
} satisfies ChartConfig;

export function SubStatusChart({ counts }: SubStatusChartProps) {
  const chartData = [
    { status: "active",    count: counts.Active,    fill: "var(--color-active)" },
    { status: "grace",     count: counts.Grace,     fill: "var(--color-grace)" },
    { status: "cancelled", count: counts.Cancelled, fill: "var(--color-cancelled)" },
    { status: "expired",   count: counts.Expired,   fill: "var(--color-expired)" },
  ];
  const total = chartData.reduce((acc, d) => acc + d.count, 0);

  return (
    <Card className="flex flex-col">
      <CardHeader className="items-center pb-0">
        <CardTitle>Subscription Status</CardTitle>
        <CardDescription>Current distribution</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        <ChartContainer
          config={chartConfig}
          className="mx-auto aspect-square max-h-[220px]"
        >
          <PieChart>
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Pie
              data={chartData}
              dataKey="count"
              nameKey="status"
              innerRadius={60}
              strokeWidth={5}
              activeIndex={0}
              activeShape={({ outerRadius = 0, ...props }: PieSectorDataItem) => (
                <Sector {...props} outerRadius={outerRadius + 10} />
              )}
            >
              <Label
                content={({ viewBox }) => {
                  if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                    return (
                      <text
                        x={viewBox.cx}
                        y={viewBox.cy}
                        textAnchor="middle"
                        dominantBaseline="middle"
                      >
                        <tspan
                          x={viewBox.cx}
                          y={viewBox.cy}
                          className="fill-foreground text-2xl font-bold"
                        >
                          {total.toLocaleString()}
                        </tspan>
                        <tspan
                          x={viewBox.cx}
                          y={(viewBox.cy || 0) + 20}
                          className="fill-muted-foreground text-xs"
                        >
                          Total Subs
                        </tspan>
                      </text>
                    );
                  }
                }}
              />
            </Pie>
          </PieChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className="flex-col gap-1 pt-2 text-sm">
        {chartData.map((d) => (
          <div key={d.status} className="flex w-full items-center justify-between">
            <div className="flex items-center gap-1.5">
              <span
                className="inline-block h-2 w-2 rounded-full"
                style={{ backgroundColor: d.fill }}
              />
              <span className="text-muted-foreground capitalize">
                {d.status === "grace" ? "Grace Period" : d.status}
              </span>
            </div>
            <span className="font-medium tabular-nums">
              {total > 0 ? ((d.count / total) * 100).toFixed(0) : 0}%
            </span>
          </div>
        ))}
      </CardFooter>
    </Card>
  );
}
