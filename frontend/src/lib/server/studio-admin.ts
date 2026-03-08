import "server-only";

import { cookies } from "next/headers";

import type { BanditMetrics, BanditStatisticsResponse } from "@/lib/bandit";
import type {
  ExperimentStudioDashboardData,
  ExperimentStudioSnapshot,
  StudioBanditHealth,
  StudioEndpointProbe,
} from "@/lib/experiment-studio";
import type {
  ExperimentLifecycleAudit,
  ExperimentSummary,
  ExperimentWinnerRecommendationAudit,
} from "@/lib/experiments";
import type { PricingTier } from "@/lib/pricing-tiers";

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

function unavailableProbe(message: string): StudioEndpointProbe {
  return { ok: false, status: null, message };
}

async function fetchProbe<T>(url: string, init?: RequestInit) {
  try {
    const res = await fetch(url, { cache: "no-store", ...init });
    const parsed = await parseResponse<T>(res);
    if (!parsed.ok) {
      return { ok: false as const, status: res.status, message: parsed.error, data: null };
    }
    return { ok: true as const, status: res.status, message: "OK", data: parsed.data };
  } catch {
    return { ok: false as const, status: null, message: "Request failed", data: null };
  }
}

export async function getAdminExperimentsFromCookies(): Promise<ExperimentSummary[] | null> {
  const token = await getAdminToken();
  if (!token) return null;

  const res = await fetchProbe<ExperimentSummary[]>(`${BACKEND_URL}/v1/admin/experiments`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.ok ? res.data : null;
}

export async function getPricingTiersFromCookies(): Promise<PricingTier[] | null> {
  const token = await getAdminToken();
  if (!token) return null;

  const res = await fetchProbe<PricingTier[]>(`${BACKEND_URL}/v1/admin/pricing-tiers`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.ok ? res.data : null;
}

async function getBanditHealth() {
  const res = await fetchProbe<StudioBanditHealth>(`${BACKEND_URL}/v1/bandit/health`);
  return {
    probe: {
      ok: res.ok,
      status: res.status,
      message: res.ok ? `${res.data.service}: ${res.data.status}` : res.message,
    },
    health: res.ok ? res.data : null,
  };
}

export async function getStudioSnapshotFromCookies(experimentId: string): Promise<ExperimentStudioSnapshot | null> {
  const experiments = await getAdminExperimentsFromCookies();
  const experiment = experiments?.find((item) => item.id === experimentId);
  if (!experiment) return null;

  const token = await getAdminToken();
  if (!token) return null;

  const lifecycleHistoryRes = await fetchProbe<ExperimentLifecycleAudit[]>(
    `${BACKEND_URL}/v1/admin/experiments/${experimentId}/lifecycle-audit`,
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  );
  const lifecycleHistory = lifecycleHistoryRes.ok ? lifecycleHistoryRes.data : [];

  const recommendationHistoryRes = await fetchProbe<ExperimentWinnerRecommendationAudit[]>(
    `${BACKEND_URL}/v1/admin/experiments/${experimentId}/winner-recommendation-audit`,
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  );
  const recommendationHistory = recommendationHistoryRes.ok ? recommendationHistoryRes.data : [];

  const banditHealthResult = await getBanditHealth();

  if (!experiment.is_bandit) {
    return {
      experiment,
      lifecycleHistory,
      recommendationHistory,
      banditHealth: banditHealthResult.health,
      banditSnapshot: null,
      endpoints: {
        statistics: unavailableProbe("Classic A/B experiment — no bandit statistics."),
        metrics: unavailableProbe("Classic A/B experiment — no bandit metrics."),
        objectives: unavailableProbe("Objective scoring is only relevant for bandit-enabled experiments."),
        windowInfo: unavailableProbe("Sliding-window runtime is only relevant for bandit-enabled experiments."),
      },
    };
  }

  const [statisticsRes, metricsRes, objectivesRes, windowInfoRes] = await Promise.all([
    fetchProbe<BanditStatisticsResponse>(
      `${BACKEND_URL}/v1/bandit/statistics?experiment_id=${experimentId}&win_probs=true`,
    ),
    fetchProbe<unknown>(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/metrics`),
    fetchProbe<Record<string, unknown>>(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/objectives`),
    fetchProbe<Record<string, unknown>>(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/window/info`),
  ]);

  return {
    experiment,
    lifecycleHistory,
    recommendationHistory,
    banditHealth: banditHealthResult.health,
    banditSnapshot: {
      experiment,
      statistics: statisticsRes.ok ? statisticsRes.data : null,
      metrics: metricsRes.ok ? normalizeMetrics(metricsRes.data) : null,
      recommendationHistory,
    },
    endpoints: {
      statistics: {
        ok: statisticsRes.ok,
        status: statisticsRes.status,
        message: statisticsRes.ok ? "Live bandit statistics available" : statisticsRes.message,
      },
      metrics: {
        ok: metricsRes.ok,
        status: metricsRes.status,
        message: metricsRes.ok ? "Advanced metrics endpoint available" : metricsRes.message,
      },
      objectives: {
        ok: objectivesRes.ok,
        status: objectivesRes.status,
        message: objectivesRes.ok ? "Objective scores endpoint available" : objectivesRes.message,
      },
      windowInfo: {
        ok: windowInfoRes.ok,
        status: windowInfoRes.status,
        message: windowInfoRes.ok ? "Sliding-window endpoint available" : windowInfoRes.message,
      },
    },
  };
}

export async function getStudioDashboardFromCookies(): Promise<ExperimentStudioDashboardData> {
  const [experiments, pricingTiers] = await Promise.all([
    getAdminExperimentsFromCookies(),
    getPricingTiersFromCookies(),
  ]);
  if (!experiments) {
    return {
      experiments: [],
      selectedExperimentId: null,
      snapshot: null,
      pricingTiers: pricingTiers ?? [],
      pricingLoadFailed: pricingTiers === null,
      loadFailed: true,
    };
  }

  const selected = pickDefaultExperiment(experiments);
  if (!selected) {
    return {
      experiments,
      selectedExperimentId: null,
      snapshot: null,
      pricingTiers: pricingTiers ?? [],
      pricingLoadFailed: pricingTiers === null,
      loadFailed: false,
    };
  }

  const snapshot = await getStudioSnapshotFromCookies(selected.id);
  return {
    experiments,
    selectedExperimentId: selected.id,
    snapshot,
    pricingTiers: pricingTiers ?? [],
    pricingLoadFailed: pricingTiers === null,
    loadFailed: false,
  };
}
