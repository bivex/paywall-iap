"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export interface UserProfile {
  user: {
    id: string;
    platform_user_id: string;
    device_id: string | null;
    platform: string;
    app_version: string;
    email: string;
    role: string;
    ltv: number;
    created_at: string;
  };
  subscriptions: {
    id: string;
    status: string;
    source: string;
    platform: string;
    product_id: string;
    plan_type: string;
    expires_at: string;
    auto_renew: boolean;
    created_at: string;
  }[];
  transactions: {
    id: string;
    amount: number;
    currency: string;
    status: string;
    provider_tx_id: string | null;
    date: string;
  }[];
  audit_log: {
    action: string;
    admin_email: string;
    detail: string;
    date: string;
  }[];
  dunning: {
    status: string;
    attempt_count: number;
    max_attempts: number;
    next_attempt_at: string | null;
    created_at: string;
  }[];
}

export async function getUserProfile(id: string): Promise<UserProfile | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return null;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/users/${id}/profile`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return null;
    return (await res.json()) as UserProfile;
  } catch {
    return null;
  }
}
