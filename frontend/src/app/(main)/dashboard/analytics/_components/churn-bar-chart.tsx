"use client";

import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import {
  Card, CardContent, CardDescription, CardHeader, CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig,
} from "@/components/ui/chart";

const data = [
  { month: "Aug", ios: 2.4, android: 2.9, web: 1.8 },
  { month: "Sep", ios: 2.2, android: 2.7, web: 1.7 },
  { month: "Oct", ios: 2.1, android: 2.6, web: 1.6 },
  { month: "Nov", ios: 2.0, android: 2.5, web: 1.6 },
  { month: "Dec", ios: 1.9, android: 2.4, web: 1.5 },
  { month: "Jan", ios: 1.8, android: 2.4, web: 1.5 },
];

const chartConfig = {
  ios:     { label: "iOS",     color: "var(--chart-1)" },
  android: { label: "Android", color: "var(--chart-2)" },
  web:     { label: "Web",     color: "var(--chart-3)" },
} satisfies ChartConfig;

export function ChurnBarChart() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Churn Rate by Platform</CardTitle>
        <CardDescription>Monthly % · last 6 months</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[220px] w-full">
          <BarChart data={data} margin={{ left: 0, right: 8 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="month" tickLine={false} axisLine={false} tickMargin={8} />
            <YAxis tickLine={false} axisLine={false} tickFormatter={(v) => `${v}%`} width={36} />
            <ChartTooltip
              content={
                <ChartTooltipContent
                  formatter={(value) => [`${value}%`, ""]}
                />
              }
            />
            <Bar dataKey="ios"     fill="var(--color-ios)"     radius={[3, 3, 0, 0]} />
            <Bar dataKey="android" fill="var(--color-android)" radius={[3, 3, 0, 0]} />
            <Bar dataKey="web"     fill="var(--color-web)"     radius={[3, 3, 0, 0]} />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
