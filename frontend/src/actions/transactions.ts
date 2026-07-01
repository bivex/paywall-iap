"use server";

import { cookies } from "next/headers";

import { serverFetch, type ServerFetchResult } from "@/lib/server-fetch";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export interface TransactionRow {
  id: string;
  amount: number;
  currency: string;
  status: "success" | "failed" | "refunded";
  provider_tx_id: string;
  receipt_hash: string;
  created_at: string;
  user_id: string;
  email: string;
  source: string;
  platform: string;
  plan_type: string;
  subscription_id: string;
}

export interface TransactionSummary {
  total_count: number;
  success_count: number;
  failed_count: number;
  refunded_count: number;
  total_revenue: number;
  total_refunded: number;
}

export interface TransactionsResponse {
  transactions: TransactionRow[];
  summary: TransactionSummary;
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface TransactionsParams {
  page?: number;
  limit?: number;
  status?: string;
  source?: string;
  platform?: string;
  search?: string;
  date_from?: string;
  date_to?: string;
}

export async function getTransactions(
  params: TransactionsParams = {},
): Promise<ServerFetchResult<TransactionsResponse>> {
  const qs = new URLSearchParams();
  if (params.page) qs.set("page", String(params.page));
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.status) qs.set("status", params.status);
  if (params.source) qs.set("source", params.source);
  if (params.platform) qs.set("platform", params.platform);
  if (params.search) qs.set("search", params.search);
  if (params.date_from) qs.set("date_from", params.date_from);
  if (params.date_to) qs.set("date_to", params.date_to);

  return serverFetch<TransactionsResponse>(`/v1/admin/transactions?${qs.toString()}`);
}

export interface TransactionDetail {
  id: string;
  amount: number;
  currency: string;
  status: string;
  provider_tx_id: string;
  receipt_hash: string;
  created_at: string;
  app_id: string;
  app_name: string;
  user: {
    id: string;
    email: string;
    platform_user_id: string;
    ltv: number;
    created_at: string;
  };
  subscription: {
    id: string;
    status: string;
    source: string;
    platform: string;
    plan_type: string;
    expires_at: string;
    created_at: string;
  };
}

export async function getTransactionDetail(id: string): Promise<TransactionDetail | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;
  const appId = cookieStore.get("admin_app_id")?.value;
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/transactions/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`,
        ...(appId ? { "X-App-ID": appId } : {}),
      },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const body = await res.json();
    // backend may wrap in { data: ... }
    return (body.data ?? body) as TransactionDetail;
  } catch {
    return null;
  }
}
