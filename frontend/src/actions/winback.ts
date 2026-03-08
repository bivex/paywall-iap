"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

import type { LaunchWinbackCampaignInput, WinbackCampaign } from "@/lib/winback";

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

export async function getWinbackCampaigns(): Promise<WinbackCampaign[] | null> {
  const token = await getAdminToken();
  if (!token) return null;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/winback-campaigns`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const parsed = await parseResponse<WinbackCampaign[]>(res);
    return parsed.ok ? parsed.data : null;
  } catch {
    return null;
  }
}

export async function launchWinbackCampaignAction(payload: LaunchWinbackCampaignInput) {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" } satisfies ActionResult<WinbackCampaign>;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/winback-campaigns`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const parsed = await parseResponse<WinbackCampaign>(res);
    if (parsed.ok) revalidatePath("/dashboard/winback");
    return parsed;
  } catch (error) {
    return { ok: false, error: String(error) } satisfies ActionResult<WinbackCampaign>;
  }
}

export async function deactivateWinbackCampaignAction(campaignId: string) {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" } satisfies ActionResult<WinbackCampaign>;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/winback-campaigns/${encodeURIComponent(campaignId)}/deactivate`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
    const parsed = await parseResponse<WinbackCampaign>(res);
    if (parsed.ok) revalidatePath("/dashboard/winback");
    return parsed;
  } catch (error) {
    return { ok: false, error: String(error) } satisfies ActionResult<WinbackCampaign>;
  }
}
