import "server-only";

import { cookies } from "next/headers";

import type { BanditMetrics, BanditSnapshot, BanditStatisticsResponse } from "@/lib/bandit";
import type { ExperimentSummary } from "@/lib/experiments";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

async function getAdminToken() {
  const cookieStore = await cookies();
  return cookieStore.get("admin_access_token")?.value;
}

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

function pickDefaultExperiment(experiments: ExperimentSummary[]) {
  return experiments.find((experiment) => experiment.status === "running") ?? experiments[0] ?? null;
}

function toNumber(value: unknown) {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function normalizeMetrics(payload: unknown): BanditMetrics {
  const raw = (payload ?? {}) as Record<string, unknown>;
  return {
    regret: toNumber(raw.regret ?? raw.Regret),
    exploration_rate: toNumber(raw.exploration_rate ?? raw.ExplorationRate),
    convergence_gap: toNumber(raw.convergence_gap ?? raw.ConvergenceGap),
    balance_index: toNumber(raw.balance_index ?? raw.BalanceIndex),
    window_utilization: toNumber(raw.window_utilization ?? raw.WindowUtilization),
    pending_rewards: toNumber(raw.pending_rewards ?? raw.PendingRewards),
  };
}

export async function getBanditExperimentsFromCookies(): Promise<ExperimentSummary[] | null> {
  const token = await getAdminToken();
  if (!token) return null;

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/experiments`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    const parsed = await parseResponse<ExperimentSummary[]>(res);
    if (!parsed.ok) return null;
    return parsed.data.filter((experiment) => experiment.is_bandit);
  } catch {
    return null;
  }
}

export async function getBanditSnapshotFromCookies(experimentId: string): Promise<BanditSnapshot | null> {
  const experiments = await getBanditExperimentsFromCookies();
  const experiment = experiments?.find((item) => item.id === experimentId);
  if (!experiment) return null;

  const [statisticsResult, metricsResult] = await Promise.allSettled([
    fetch(`${BACKEND_URL}/v1/bandit/statistics?experiment_id=${experimentId}&win_probs=true`, {
      cache: "no-store",
    }),
    fetch(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/metrics`, {
      cache: "no-store",
    }),
  ]);

  let statistics: BanditStatisticsResponse | null = null;
  if (statisticsResult.status === "fulfilled") {
    const parsed = await parseResponse<BanditStatisticsResponse>(statisticsResult.value);
    if (parsed.ok) statistics = parsed.data;
  }

  let metrics: BanditMetrics | null = null;
  if (metricsResult.status === "fulfilled") {
    const parsed = await parseResponse<BanditMetrics>(metricsResult.value);
    if (parsed.ok) metrics = normalizeMetrics(parsed.data);
  }

  return { experiment, statistics, metrics };
}

export async function getBanditDashboardFromCookies() {
  const experiments = await getBanditExperimentsFromCookies();
  if (!experiments) {
    return { experiments: [], selectedExperimentId: null, snapshot: null, loadFailed: true };
  }

  const selected = pickDefaultExperiment(experiments);
  if (!selected) {
    return { experiments, selectedExperimentId: null, snapshot: null, loadFailed: false };
  }

  const snapshot = await getBanditSnapshotFromCookies(selected.id);
  return {
    experiments,
    selectedExperimentId: selected.id,
    snapshot,
    loadFailed: false,
  };
}
