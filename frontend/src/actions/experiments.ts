"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

import type { ExperimentInput, ExperimentSummary } from "@/lib/experiments";

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

export async function getExperiments(): Promise<ExperimentSummary[] | null> {
  const token = await getAdminToken();
  if (!token) return null;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/experiments`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const parsed = await parseResponse<ExperimentSummary[]>(res);
    return parsed.ok ? parsed.data : null;
  } catch {
    return null;
  }
}

export async function createExperimentAction(payload: ExperimentInput) {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" } satisfies ActionResult<ExperimentSummary>;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/experiments`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const parsed = await parseResponse<ExperimentSummary>(res);
    if (parsed.ok) revalidatePath("/dashboard/experiments");
    return parsed;
  } catch (error) {
    return { ok: false, error: String(error) } satisfies ActionResult<ExperimentSummary>;
  }
}

async function postExperimentLifecycleAction(id: string, action: "pause" | "resume" | "complete") {
  const token = await getAdminToken();
  if (!token) return { ok: false, error: "Unauthorized" } satisfies ActionResult<ExperimentSummary>;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/experiments/${id}/${action}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    });
    const parsed = await parseResponse<ExperimentSummary>(res);
    if (parsed.ok) revalidatePath("/dashboard/experiments");
    return parsed;
  } catch (error) {
    return { ok: false, error: String(error) } satisfies ActionResult<ExperimentSummary>;
  }
}

export async function pauseExperimentAction(id: string) {
  return postExperimentLifecycleAction(id, "pause");
}

export async function resumeExperimentAction(id: string) {
  return postExperimentLifecycleAction(id, "resume");
}

export async function completeExperimentAction(id: string) {
  return postExperimentLifecycleAction(id, "complete");
}
