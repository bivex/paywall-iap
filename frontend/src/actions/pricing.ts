"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

import type { PricingTier, PricingTierInput } from "@/lib/pricing-tiers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

type ActionResult<T> = { ok: true; data: T } | { ok: false; error: string };

async function getAdminToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get("admin_access_token")?.value;
}

async function parseResponse<T>(res: Response): Promise<ActionResult<T>> {
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    return {
      ok: false,
      error:
        (body as { message?: string; error?: string }).message ??
        (body as { error?: string }).error ??
        `HTTP ${res.status}`,
    };
  }
  return { ok: true, data: ((body as { data?: T }).data ?? body) as T };
}

async function mutate<T>(path: string, method: "POST" | "PUT", payload?: PricingTierInput): Promise<ActionResult<T>> {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" };

  try {
    const res = await fetch(`${BACKEND_URL}${path}`, {
      method,
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: payload ? JSON.stringify(payload) : undefined,
    });
    const parsed = await parseResponse<T>(res);
    if (parsed.ok) {
      revalidatePath("/dashboard/pricing");
      revalidatePath("/dashboard/experiments/studio");
    }
    return parsed;
  } catch (error) {
    return { ok: false, error: String(error) };
  }
}

export async function getPricingTiers(): Promise<PricingTier[] | null> {
  const token = await getAdminToken();
  if (!token) return null;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/pricing-tiers`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const parsed = await parseResponse<PricingTier[]>(res);
    return parsed.ok ? parsed.data : null;
  } catch {
    return null;
  }
}

export async function createPricingTierAction(payload: PricingTierInput) {
  return mutate<PricingTier>("/v1/admin/pricing-tiers", "POST", payload);
}

export async function updatePricingTierAction(id: string, payload: PricingTierInput) {
  return mutate<PricingTier>(`/v1/admin/pricing-tiers/${id}`, "PUT", payload);
}

export async function activatePricingTierAction(id: string) {
  return mutate<PricingTier>(`/v1/admin/pricing-tiers/${id}/activate`, "POST");
}

export async function deactivatePricingTierAction(id: string) {
  return mutate<PricingTier>(`/v1/admin/pricing-tiers/${id}/deactivate`, "POST");
}
