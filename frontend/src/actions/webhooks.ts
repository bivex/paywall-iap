"use server";

import { serverFetch, type ServerFetchResult } from "@/lib/server-fetch";

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

export async function getWebhooks(params: WebhooksParams = {}): Promise<ServerFetchResult<WebhooksResponse>> {
  const qs = new URLSearchParams();
  if (params.page) qs.set("page", String(params.page));
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.provider) qs.set("provider", params.provider);
  if (params.status) qs.set("status", params.status);
  if (params.search) qs.set("search", params.search);
  if (params.date_from) qs.set("date_from", params.date_from);
  if (params.date_to) qs.set("date_to", params.date_to);

  return serverFetch<WebhooksResponse>(`/v1/admin/webhooks?${qs.toString()}`);
}
