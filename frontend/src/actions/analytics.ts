"use server";

import { serverFetch, type ServerFetchResult } from "@/lib/server-fetch";

export interface TrendPoint {
  month: string;
  mrr: number;
  active_count: number;
  new_subs: number;
}

export interface PlatformRow {
  platform: string;
  count: number;
  mrr: number;
}

export interface PlanRow {
  plan_type: string;
  count: number;
  mrr: number;
}

export interface AnalyticsReport {
  mrr: number;
  arr: number;
  ltv: number;
  total_revenue: number;
  churn_rate: number;
  new_subs_month: number;
  trend: TrendPoint[];
  by_platform: PlatformRow[];
  by_plan: PlanRow[];
  status_counts: {
    active: number;
    grace: number;
    cancelled: number;
    expired: number;
  };
}

export async function getAnalyticsReport(): Promise<ServerFetchResult<AnalyticsReport>> {
  return serverFetch<AnalyticsReport>("/v1/admin/analytics/report");
}
