"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export interface AppPaywall {
  id: string;
  app_id: string;
  name: string;
  description: string;
  definition: Record<string, unknown>;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PaywallsResponse {
  paywalls: AppPaywall[];
  total: number;
}

async function getHeaders(): Promise<Record<string, string>> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  const appId = cookieStore.get("admin_app_id")?.value;
  const h: Record<string, string> = { "Content-Type": "application/json" };
  if (token) h["Authorization"] = `Bearer ${token}`;
  if (appId) h["X-App-ID"] = appId;
  return h;
}

export async function getPaywalls(): Promise<PaywallsResponse | null> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/paywalls`, {
      headers: await getHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return null;
    const body = await res.json();
    return (body.data ?? body) as PaywallsResponse;
  } catch {
    return null;
  }
}

export async function getPaywall(id: string): Promise<AppPaywall | null> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/paywalls/${id}`, {
      headers: await getHeaders(),
      cache: "no-store",
    });
    if (!res.ok) return null;
    const body = await res.json();
    return (body.data ?? body) as AppPaywall;
  } catch {
    return null;
  }
}

export async function createPaywall(payload: {
  name: string;
  description?: string;
  definition: Record<string, unknown>;
  is_active?: boolean;
}): Promise<{ ok: true; data: AppPaywall } | { ok: false; error: string }> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/paywalls`, {
      method: "POST",
      headers: await getHeaders(),
      body: JSON.stringify(payload),
    });
    const body = await res.json();
    if (!res.ok) return { ok: false, error: body.error ?? body.message ?? `HTTP ${res.status}` };
    return { ok: true, data: (body.data ?? body) as AppPaywall };
  } catch (e) {
    return { ok: false, error: String(e) };
  }
}

export async function updatePaywall(
  id: string,
  payload: {
    name: string;
    description?: string;
    definition: Record<string, unknown>;
    is_active?: boolean;
  }
): Promise<{ ok: true; data: AppPaywall } | { ok: false; error: string }> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/paywalls/${id}`, {
      method: "PUT",
      headers: await getHeaders(),
      body: JSON.stringify(payload),
    });
    const body = await res.json();
    if (!res.ok) return { ok: false, error: body.error ?? body.message ?? `HTTP ${res.status}` };
    return { ok: true, data: (body.data ?? body) as AppPaywall };
  } catch (e) {
    return { ok: false, error: String(e) };
  }
}

export async function activatePaywall(
  id: string
): Promise<{ ok: true; data: AppPaywall } | { ok: false; error: string }> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/paywalls/${id}/activate`, {
      method: "POST",
      headers: await getHeaders(),
    });
    const body = await res.json();
    if (!res.ok) return { ok: false, error: body.error ?? body.message ?? `HTTP ${res.status}` };
    return { ok: true, data: (body.data ?? body) as AppPaywall };
  } catch (e) {
    return { ok: false, error: String(e) };
  }
}

export async function deletePaywall(
  id: string
): Promise<{ ok: boolean; error?: string }> {
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/paywalls/${id}`, {
      method: "DELETE",
      headers: await getHeaders(),
    });
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      return { ok: false, error: (body as { error?: string }).error ?? `HTTP ${res.status}` };
    }
    return { ok: true };
  } catch (e) {
    return { ok: false, error: String(e) };
  }
}
