"use client";

import * as React from "react";
import { Label, Pie, PieChart } from "recharts";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import type { PlanRow } from "@/actions/analytics";

const COLORS = ["var(--chart-1)", "var(--chart-2)", "var(--chart-3)"];

interface Props { data: PlanRow[]; totalMrr: number }

export function RevenueDonutChart({ data, totalMrr }: Props) {
  const chartData = data.map((d, i) => ({
    name: d.plan_type,
    value: d.mrr,
    fill: COLORS[i % COLORS.length],
  }));

  const config: ChartConfig = Object.fromEntries(
    data.map((d, i) => [d.plan_type, { label: d.plan_type, color: COLORS[i % COLORS.length] }])
  );
  config.value = { label: "MRR" };

  return (
    <Card className="flex flex-col">
      <CardHeader className="items-center pb-0">
        <CardTitle>MRR by Plan</CardTitle>
        <CardDescription>Active + grace subscriptions</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        <ChartContainer config={config} className="mx-auto aspect-square max-h-[220px]">
          <PieChart>
            <ChartTooltip cursor={false} content={<ChartTooltipContent formatter={(v) => [`$${Number(v).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`, ""]} hideLabel />} />
            <Pie data={chartData} dataKey="value" nameKey="name" innerRadius={58} strokeWidth={4}>
              <Label content={({ viewBox }) => {
                if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                  return (
                    <text x={viewBox.cx} y={viewBox.cy} textAnchor="middle" dominantBaseline="middle">
                      <tspan x={viewBox.cx} y={viewBox.cy} className="fill-foreground text-xl font-bold">
                        ${totalMrr.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                      </tspan>
                      <tspan x={viewBox.cx} y={(viewBox.cy ?? 0) + 18} className="fill-muted-foreground text-xs">MRR</tspan>
                    </text>
                  );
                }
              }} />
            </Pie>
          </PieChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className="flex-col gap-1 text-sm pt-2">
        {data.map((d, i) => (
          <div key={d.plan_type} className="flex items-center justify-between w-full text-xs">
            <span className="flex items-center gap-1.5">
              <span className="h-2 w-2 rounded-full" style={{ background: COLORS[i % COLORS.length] }} />
              {d.plan_type} ({d.count} subs)
            </span>
            <span className="font-medium">${d.mrr.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
          </div>
        ))}
      </CardFooter>
    </Card>
  );
}
