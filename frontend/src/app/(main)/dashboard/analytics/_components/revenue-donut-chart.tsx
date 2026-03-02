"use client";

import * as React from "react";
import { TrendingUp } from "lucide-react";
import { Label, Pie, PieChart } from "recharts";
import {
  Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig,
} from "@/components/ui/chart";

const chartData = [
  { plan: "pro_annual",     revenue: 7200, fill: "var(--color-pro_annual)" },
  { plan: "basic_monthly",  revenue: 3100, fill: "var(--color-basic_monthly)" },
  { plan: "enterprise",     revenue: 2150, fill: "var(--color-enterprise)" },
];

const chartConfig = {
  revenue: { label: "Revenue" },
  pro_annual:    { label: "Pro Annual",    color: "var(--chart-1)" },
  basic_monthly: { label: "Basic Monthly", color: "var(--chart-2)" },
  enterprise:    { label: "Enterprise",    color: "var(--chart-3)" },
} satisfies ChartConfig;

export function RevenueDonutChart() {
  const total = React.useMemo(
    () => chartData.reduce((s, d) => s + d.revenue, 0),
    []
  );

  return (
    <Card className="flex flex-col">
      <CardHeader className="items-center pb-0">
        <CardTitle>Revenue by Plan</CardTitle>
        <CardDescription>January 2026</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        <ChartContainer config={chartConfig} className="mx-auto aspect-square max-h-[240px]">
          <PieChart>
            <ChartTooltip
              cursor={false}
              content={
                <ChartTooltipContent
                  formatter={(value) => [`$${Number(value).toLocaleString("en-US")}`, ""]}
                  hideLabel
                />
              }
            />
            <Pie data={chartData} dataKey="revenue" nameKey="plan" innerRadius={64} strokeWidth={5}>
              <Label
                content={({ viewBox }) => {
                  if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                    return (
                      <text x={viewBox.cx} y={viewBox.cy} textAnchor="middle" dominantBaseline="middle">
                        <tspan x={viewBox.cx} y={viewBox.cy} className="fill-foreground text-2xl font-bold">
                          ${total.toLocaleString("en-US")}
                        </tspan>
                        <tspan x={viewBox.cx} y={(viewBox.cy ?? 0) + 20} className="fill-muted-foreground text-xs">
                          MRR
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
      <CardFooter className="flex-col gap-1 text-sm">
        <div className="flex items-center gap-2 font-medium leading-none">
          MRR up 8.3% this month <TrendingUp className="h-4 w-4 text-green-500" />
        </div>
        <div className="text-muted-foreground leading-none">Pro Annual drives 58% of revenue</div>
      </CardFooter>
    </Card>
  );
}
