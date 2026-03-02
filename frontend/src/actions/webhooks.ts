"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export interface WebhookEvent {
  id: string;
  provider: string;
  event_type: string;
  event_id: string;
  processed: boolean;
  processed_at: string | null;
  created_at: string;
}

export interface WebhookSummary {
  total: number;
  pending: number;
  processed: number;
}

export interface WebhooksResponse {
  webhooks: WebhookEvent[];
  summary: WebhookSummary;
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface WebhooksParams {
  page?: number;
  limit?: number;
  provider?: string;
  status?: string;
  search?: string;
  date_from?: string;
  date_to?: string;
}

export async function getWebhooks(params: WebhooksParams = {}): Promise<WebhooksResponse | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;

  const qs = new URLSearchParams();
  if (params.page) qs.set("page", String(params.page));
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.provider) qs.set("provider", params.provider);
  if (params.status) qs.set("status", params.status);
  if (params.search) qs.set("search", params.search);
  if (params.date_from) qs.set("date_from", params.date_from);
  if (params.date_to) qs.set("date_to", params.date_to);

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/webhooks?${qs.toString()}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return null;
    return (await res.json()) as WebhooksResponse;
  } catch {
    return null;
  }
}
