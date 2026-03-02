"use server";

import { cookies } from "next/headers";
import { revalidatePath } from "next/cache";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

async function adminFetch(path: string, body: object) {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return { ok: false, error: "Unauthorized" };

  try {
    const res = await fetch(`${BACKEND_URL}${path}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const d = await res.json().catch(() => ({}));
      return { ok: false, error: (d as any).message ?? `HTTP ${res.status}` };
    }
    const data = await res.json().catch(() => ({}));
    return { ok: true, data };
  } catch (e) {
    return { ok: false, error: String(e) };
  }
}

export async function forceCancelAction(userId: string, reason: string) {
  const result = await adminFetch(`/v1/admin/users/${userId}/force-cancel`, { reason });
  if (result.ok) revalidatePath(`/dashboard/users/${userId}`);
  return result;
}

export async function forceRenewAction(userId: string, days: number, reason: string) {
  const result = await adminFetch(`/v1/admin/users/${userId}/force-renew`, { days, reason });
  if (result.ok) revalidatePath(`/dashboard/users/${userId}`);
  return result;
}

export async function grantGraceAction(userId: string, days: number, reason: string) {
  const result = await adminFetch(`/v1/admin/users/${userId}/grant-grace`, { days, reason });
  if (result.ok) revalidatePath(`/dashboard/users/${userId}`);
  return result;
}
