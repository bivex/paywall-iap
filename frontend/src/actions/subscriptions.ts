"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export interface SubscriptionRow {
  id: string;
  status: string; // active | grace | cancelled | expired
  source: string; // iap | stripe | paddle
  platform: string; // ios | android | web
  plan_type: string; // monthly | annual | lifetime
  expires_at: string; // ISO8601
  created_at: string; // ISO8601
  user_id: string;
  email: string;
  ltv: number;
}

export interface SubscriptionsResponse {
  subscriptions: SubscriptionRow[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface SubscriptionsParams {
  page?: number;
  limit?: number;
  status?: string;
  source?: string;
  platform?: string;
  plan_type?: string;
  search?: string;
  date_from?: string;
  date_to?: string;
}

export async function getSubscriptions(
  params: SubscriptionsParams = {},
): Promise<SubscriptionsResponse | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;

  const qs = new URLSearchParams();
  if (params.page) qs.set("page", String(params.page));
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.status) qs.set("status", params.status);
  if (params.source) qs.set("source", params.source);
  if (params.platform) qs.set("platform", params.platform);
  if (params.plan_type) qs.set("plan_type", params.plan_type);
  if (params.search) qs.set("search", params.search);
  if (params.date_from) qs.set("date_from", params.date_from);
  if (params.date_to) qs.set("date_to", params.date_to);

  try {
    const res = await fetch(
      `${BACKEND_URL}/v1/admin/subscriptions?${qs.toString()}`,
      {
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      },
    );
    if (!res.ok) return null;
    return (await res.json()) as SubscriptionsResponse;
  } catch {
    return null;
  }
}


