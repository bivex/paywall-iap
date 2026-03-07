import "server-only";

import type {
  MultiObjectiveDashboardData,
  MultiObjectiveSnapshot,
  ObjectiveEndpointProbe,
  ObjectiveScoreEntry,
  ObjectiveScoresByArm,
  ObjectiveServiceHealth,
} from "@/lib/multi-objective";
import { getBanditExperimentsFromCookies, getBanditSnapshotFromCookies } from "@/lib/server/bandit-admin";

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

function manualProbe(message: string): ObjectiveEndpointProbe {
  return { state: "manual", status: null, message };
}

function toNumber(value: unknown) {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function normalizeObjectiveScoreEntry(objectiveType: string, payload: unknown): ObjectiveScoreEntry {
  const raw = (payload ?? {}) as Record<string, unknown>;
  return {
    objectiveType:
      (typeof raw.objectiveType === "string" && raw.objectiveType) ||
      (typeof raw.ObjectiveType === "string" && raw.ObjectiveType) ||
      objectiveType,
    score: toNumber(raw.score ?? raw.Score),
    alpha: toNumber(raw.alpha ?? raw.Alpha),
    beta: toNumber(raw.beta ?? raw.Beta),
    samples: toNumber(raw.samples ?? raw.Samples),
    conversions: toNumber(raw.conversions ?? raw.Conversions),
    revenue: toNumber(raw.revenue ?? raw.Revenue),
    avgLtv: toNumber(raw.avgLtv ?? raw.AvgLTV ?? raw.avg_ltv),
  };
}

function normalizeObjectiveScores(payload: unknown): ObjectiveScoresByArm {
  const raw = (payload ?? {}) as Record<string, unknown>;
  const result: ObjectiveScoresByArm = {};

  for (const [armId, armPayload] of Object.entries(raw)) {
    if (!armPayload || typeof armPayload !== "object") continue;
    const armScores: Record<string, ObjectiveScoreEntry> = {};
    for (const [objectiveKey, objectivePayload] of Object.entries(armPayload as Record<string, unknown>)) {
      armScores[objectiveKey] = normalizeObjectiveScoreEntry(objectiveKey, objectivePayload);
    }
    result[armId] = armScores;
  }

  return result;
}

async function fetchProbe<T>(url: string): Promise<{ probe: ObjectiveEndpointProbe; data: T | null }> {
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
  const result = await fetchProbe<ObjectiveServiceHealth>(`${BACKEND_URL}/v1/bandit/health`);
  return result.data;
}

export async function getMultiObjectiveSnapshotFromCookies(
  experimentId: string,
): Promise<MultiObjectiveSnapshot | null> {
  const experiments = await getBanditExperimentsFromCookies();
  const experiment = experiments?.find((item) => item.id === experimentId);
  if (!experiment) return null;

  const [banditSnapshot, serviceHealth, objectiveScoresResult] = await Promise.all([
    getBanditSnapshotFromCookies(experimentId),
    getServiceHealth(),
    fetchProbe<unknown>(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/objectives`),
  ]);

  return {
    experiment,
    banditSnapshot,
    serviceHealth,
    objectiveScores: objectiveScoresResult.data ? normalizeObjectiveScores(objectiveScoresResult.data) : null,
    probes: {
      objectiveScores: objectiveScoresResult.probe,
      objectiveConfig: manualProbe(
        "Live PUT route exists, but this page does not auto-call objective configuration mutations during a read-only load.",
      ),
    },
  };
}

export async function getMultiObjectiveDashboardFromCookies(): Promise<MultiObjectiveDashboardData> {
  const experiments = await getBanditExperimentsFromCookies();
  if (!experiments) {
    return { experiments: [], selectedExperimentId: null, snapshot: null, loadFailed: true };
  }

  const selected = pickDefaultExperiment(experiments);
  if (!selected) {
    return { experiments, selectedExperimentId: null, snapshot: null, loadFailed: false };
  }

  const snapshot = await getMultiObjectiveSnapshotFromCookies(selected.id);
  return {
    experiments,
    selectedExperimentId: selected.id,
    snapshot,
    loadFailed: false,
  };
}
