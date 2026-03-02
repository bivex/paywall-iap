"use server";

import { cookies } from "next/headers";

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

export async function getAnalyticsReport(): Promise<AnalyticsReport | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;

  const base = process.env.BACKEND_URL ?? "http://api:8080";
  try {
    const res = await fetch(`${base}/v1/admin/analytics/report`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return null;
    return res.json() as Promise<AnalyticsReport>;
  } catch {
    return null;
  }
}
