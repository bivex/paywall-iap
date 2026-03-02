/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-02 05:53
 * Last Updated: 2026-03-02 05:53
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

"use server";

import { cookies } from "next/headers";

export interface DunningRow {
  id: string;
  email: string;
  user_id: string;
  plan_type: string;
  status: string;
  attempt_count: number;
  max_attempts: number;
  next_attempt_at: string | null;
  last_attempt_at: string | null;
  created_at: string;
}

export interface DunningStats {
  pending: number;
  in_progress: number;
  recovered: number;
  failed: number;
}

export interface WebhookRow {
  id: string;
  provider: string;
  event_type: string;
  event_id: string;
  processed: boolean;
  processed_at: string | null;
  created_at: string;
}

export interface WebhookProviderStat {
  provider: string;
  total: number;
  processed: number;
}

export interface MatomoStats {
  pending: number;
  processing: number;
  sent: number;
  failed: number;
  total: number;
}

export interface RevenueOpsReport {
  dunning: {
    queue: DunningRow[];
    stats: DunningStats;
  };
  webhooks: {
    events: WebhookRow[];
    total: number;
    unprocessed: number;
    by_provider: WebhookProviderStat[];
    page: number;
    page_size: number;
    total_pages: number;
  };
  matomo: {
    stats: MatomoStats;
  };
}

export async function getRevenueOps(whPage = 1): Promise<RevenueOpsReport | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;

  const base = process.env.BACKEND_URL ?? "http://api:8080";
  try {
    const res = await fetch(
      `${base}/v1/admin/revenue-ops?wh_page=${whPage}&wh_page_size=20`,
      {
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      }
    );
    if (!res.ok) return null;
    return res.json() as Promise<RevenueOpsReport>;
  } catch {
    return null;
  }
}
