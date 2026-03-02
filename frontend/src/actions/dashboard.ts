/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-02 03:41
 * Last Updated: 2026-03-02 03:41
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export interface MonthlyMRR {
  Month: string; // "2025-09"
  MRR: number;
}

export interface SubscriptionStatusCounts {
  Active: number;
  Grace: number;
  Cancelled: number;
  Expired: number;
}

export interface WebhookProviderHealth {
  Provider: string;
  Unprocessed: number;
  Total: number;
}

export interface AuditLogEntry {
  Time: string; // ISO8601
  Action: string;
  Detail: string;
}

export interface DashboardMetrics {
  active_users: number;
  active_subs: number;
  mrr: number;
  arr: number;
  churn_risk: number;
  mrr_trend: MonthlyMRR[];
  status_counts: SubscriptionStatusCounts;
  audit_log: AuditLogEntry[];
  webhook_health: WebhookProviderHealth[];
  last_updated: string;
}

export async function getDashboardMetrics(): Promise<DashboardMetrics | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/dashboard/metrics`, {
      headers: { Authorization: `Bearer ${token}` },
      next: { revalidate: 60 }, // cache for 60 s, revalidate in background
    });
    if (!res.ok) return null;
    return (await res.json()) as DashboardMetrics;
  } catch {
    return null;
  }
}
