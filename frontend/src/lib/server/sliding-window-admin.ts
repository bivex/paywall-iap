import "server-only";

import { getBanditExperimentsFromCookies, getBanditSnapshotFromCookies } from "@/lib/server/bandit-admin";
import type {
  SlidingWindowDashboardData,
  SlidingWindowEndpointProbe,
  SlidingWindowServiceHealth,
  SlidingWindowSnapshot,
} from "@/lib/sliding-window";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

async function parseResponse<T>(res: Response): Promise<{ ok: true; data: T } | { ok: false; error: string }> {
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

function pickDefaultExperiment<T extends { status: string }>(items: T[]) {
  return items.find((item) => item.status === "running") ?? items[0] ?? null;
}

function manualProbe(message: string): SlidingWindowEndpointProbe {
  return { state: "manual", status: null, message };
}

async function fetchProbe<T>(url: string): Promise<{ probe: SlidingWindowEndpointProbe; data: T | null }> {
  try {
    const res = await fetch(url, { cache: "no-store" });
    const parsed = await parseResponse<T>(res);
    if (!parsed.ok) {
      return {
        probe: { state: "unavailable", status: res.status, message: parsed.error },
        data: null,
      };
    }

    return {
      probe: { state: "available", status: res.status, message: "OK" },
      data: parsed.data,
    };
  } catch {
    return {
      probe: { state: "unavailable", status: null, message: "Request failed" },
      data: null,
    };
  }
}

async function getServiceHealth() {
  const result = await fetchProbe<SlidingWindowServiceHealth>(`${BACKEND_URL}/v1/bandit/health`);
  return result.data;
}

export async function getSlidingWindowSnapshotFromCookies(experimentId: string): Promise<SlidingWindowSnapshot | null> {
  const experiments = await getBanditExperimentsFromCookies();
  const experiment = experiments?.find((item) => item.id === experimentId);
  if (!experiment) return null;

  const [banditSnapshot, serviceHealth, windowInfoResult, windowEventsResult] = await Promise.all([
    getBanditSnapshotFromCookies(experimentId),
    getServiceHealth(),
    fetchProbe<Record<string, unknown>>(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/window/info`),
    fetchProbe<Record<string, unknown>>(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/window/events`),
  ]);

  return {
    experiment,
    banditSnapshot,
    serviceHealth,
    probes: {
      windowInfo: windowInfoResult.probe,
      windowEvents: windowEventsResult.probe,
      trimWindow: manualProbe("Live POST route exists. Use the explicit trim control on this page to call it."),
    },
  };
}

export async function getSlidingWindowDashboardFromCookies(): Promise<SlidingWindowDashboardData> {
  const experiments = await getBanditExperimentsFromCookies();
  if (!experiments) {
    return { experiments: [], selectedExperimentId: null, snapshot: null, loadFailed: true };
  }

  const selected = pickDefaultExperiment(experiments);
  if (!selected) {
    return { experiments, selectedExperimentId: null, snapshot: null, loadFailed: false };
  }

  const snapshot = await getSlidingWindowSnapshotFromCookies(selected.id);
  return {
    experiments,
    selectedExperimentId: selected.id,
    snapshot,
    loadFailed: false,
  };
}
