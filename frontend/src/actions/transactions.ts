"use server";

import { cookies } from "next/headers";

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
): Promise<TransactionsResponse | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;

  const qs = new URLSearchParams();
  if (params.page) qs.set("page", String(params.page));
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.status) qs.set("status", params.status);
  if (params.source) qs.set("source", params.source);
  if (params.platform) qs.set("platform", params.platform);
  if (params.search) qs.set("search", params.search);
  if (params.date_from) qs.set("date_from", params.date_from);
  if (params.date_to) qs.set("date_to", params.date_to);

  try {
    const res = await fetch(
      `${BACKEND_URL}/v1/admin/transactions?${qs.toString()}`,
      {
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      },
    );
    if (!res.ok) return null;
    return (await res.json()) as TransactionsResponse;
  } catch {
    return null;
  }
}

export interface TransactionDetail {
  id: string;
  amount: number;
  currency: string;
  status: string;
  provider_tx_id: string;
  receipt_hash: string;
  created_at: string;
  user: {
    id: string;
    email: string;
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
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/transactions/${id}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return null;
    return (await res.json()) as TransactionDetail;
  } catch {
    return null;
  }
}
