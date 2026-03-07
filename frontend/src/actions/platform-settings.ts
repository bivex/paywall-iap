"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

import { DEFAULT_PLATFORM_SETTINGS, type PlatformSettings } from "@/lib/platform-settings";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

function cloneDefaults(): PlatformSettings {
  return JSON.parse(JSON.stringify(DEFAULT_PLATFORM_SETTINGS)) as PlatformSettings;
}

async function getAdminToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  return cookieStore.get("admin_access_token")?.value;
}

async function parseResponse<T>(res: Response): Promise<{ ok: boolean; data?: T; error?: string }> {
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

export async function getPlatformSettings(): Promise<PlatformSettings> {
  const token = await getAdminToken();
  if (!token) return cloneDefaults();

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/settings`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const parsed = await parseResponse<PlatformSettings>(res);
    return parsed.ok && parsed.data ? parsed.data : cloneDefaults();
  } catch {
    return cloneDefaults();
  }
}

export async function updatePlatformSettings(payload: PlatformSettings) {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" };

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/settings`, {
      method: "PUT",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const parsed = await parseResponse<PlatformSettings>(res);
    if (parsed.ok) revalidatePath("/dashboard/settings");
    return parsed;
  } catch (error) {
    return { ok: false, error: String(error) };
  }
}

export async function changeAdminPasswordAction(input: {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}) {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" };

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/settings/password`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({
        current_password: input.currentPassword,
        new_password: input.newPassword,
        confirm_password: input.confirmPassword,
      }),
    });
    return await parseResponse<{ ok: boolean }>(res);
  } catch (error) {
    return { ok: false, error: String(error) };
  }
}
