import "server-only";

import { cookies } from "next/headers";

import type { App } from "@/stores/app-store";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

async function getAdminToken() {
  const cookieStore = await cookies();
  return cookieStore.get("admin_access_token")?.value;
}

async function adminFetch<T>(path: string, init?: RequestInit): Promise<T | null> {
  const token = await getAdminToken();
  if (!token) return null;
  try {
    const res = await fetch(`${BACKEND_URL}${path}`, {
      cache: "no-store",
      ...init,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
        ...(init?.headers ?? {}),
      },
    });
    if (!res.ok) return null;
    const body = await res.json();
    return (body?.data ?? body) as T;
  } catch {
    return null;
  }
}

export async function getAppsFromServer(): Promise<App[]> {
  const result = await adminFetch<{ apps: App[] }>("/v1/admin/apps");
  return result?.apps ?? [];
}

export async function createAppOnServer(
  name: string,
  bundleId: string,
  platform: string,
  displayName?: string,
): Promise<App | null> {
  return adminFetch<App>("/v1/admin/apps", {
    method: "POST",
    body: JSON.stringify({ name, bundle_id: bundleId, platform, display_name: displayName ?? name }),
  });
}

export async function updateAppOnServer(id: string, patch: Partial<App>): Promise<App | null> {
  return adminFetch<App>(`/v1/admin/apps/${id}`, {
    method: "PUT",
    body: JSON.stringify(patch),
  });
}

export async function deleteAppOnServer(id: string): Promise<boolean> {
  const token = await getAdminToken();
  if (!token) return false;
  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/apps/${id}`, {
      method: "DELETE",
      cache: "no-store",
      headers: { Authorization: `Bearer ${token}` },
    });
    return res.status === 204;
  } catch {
    return false;
  }
}
