"use client";

import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from "@/components/ui/chart";
import type { PlatformRow } from "@/actions/analytics";

const chartConfig: ChartConfig = {
  count: { label: "Active Subs", color: "var(--chart-1)" },
  mrr:   { label: "MRR (USD)",   color: "var(--chart-2)" },
};

interface Props { data: PlatformRow[] }

export function PlatformBarChart({ data }: Props) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Active Subscriptions by Platform</CardTitle>
        <CardDescription>Current active + grace · real data</CardDescription>
      </CardHeader>
      <CardContent>
        <ChartContainer config={chartConfig} className="h-[220px] w-full">
          <BarChart data={data} margin={{ left: 0, right: 8 }}>
            <CartesianGrid vertical={false} />
            <XAxis dataKey="platform" tickLine={false} axisLine={false} tickMargin={8} />
            <YAxis yAxisId="left"  tickLine={false} axisLine={false} width={30} />
            <YAxis yAxisId="right" orientation="right" tickLine={false} axisLine={false} tickFormatter={(v) => `$${v}`} width={48} />
            <ChartTooltip content={<ChartTooltipContent />} />
            <Bar yAxisId="left"  dataKey="count" fill="var(--chart-1)" radius={[4,4,0,0]} />
            <Bar yAxisId="right" dataKey="mrr"   fill="var(--chart-2)" radius={[4,4,0,0]} />
          </BarChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}
