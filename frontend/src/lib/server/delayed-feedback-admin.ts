import "server-only";

import type {
  DelayedEndpointProbe,
  DelayedFeedbackDashboardData,
  DelayedFeedbackServiceHealth,
  DelayedFeedbackSnapshot,
} from "@/lib/delayed-feedback";
import { getBanditExperimentsFromCookies, getBanditSnapshotFromCookies } from "@/lib/server/bandit-admin";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";
const PROBE_UUID = "11111111-1111-1111-1111-111111111111";

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

function manualProbe(message: string): DelayedEndpointProbe {
  return { state: "manual", status: null, message };
}

interface ProbeOptions {
  acceptedErrorStatuses?: number[];
  acceptedStatusMessages?: Partial<Record<number, string>>;
}

async function fetchProbe<T>(
  url: string,
  options: ProbeOptions = {},
): Promise<{ probe: DelayedEndpointProbe; data: T | null }> {
  try {
    const res = await fetch(url, { cache: "no-store" });
    const parsed = await parseResponse<T>(res);
    if (!parsed.ok) {
      if (options.acceptedErrorStatuses?.includes(res.status)) {
        return {
          probe: {
            state: "available",
            status: res.status,
            message: options.acceptedStatusMessages?.[res.status] ?? parsed.error,
          },
          data: null,
        };
      }

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
  const result = await fetchProbe<DelayedFeedbackServiceHealth>(`${BACKEND_URL}/v1/bandit/health`);
  return {
    probe: result.probe,
    health: result.data,
  };
}

export async function getDelayedFeedbackSnapshotFromCookies(
  experimentId: string,
): Promise<DelayedFeedbackSnapshot | null> {
  const experiments = await getBanditExperimentsFromCookies();
  const experiment = experiments?.find((item) => item.id === experimentId);
  if (!experiment) return null;

  const [banditSnapshot, healthResult, pendingById, userPending] = await Promise.all([
    getBanditSnapshotFromCookies(experimentId),
    getServiceHealth(),
    fetchProbe<Record<string, unknown>>(`${BACKEND_URL}/v1/bandit/pending/${PROBE_UUID}`, {
      acceptedErrorStatuses: [404],
      acceptedStatusMessages: {
        404: "Endpoint reachable; sentinel pending reward was not found, which is expected for this read-only probe.",
      },
    }),
    fetchProbe<Record<string, unknown>>(`${BACKEND_URL}/v1/bandit/users/${PROBE_UUID}/pending`),
  ]);

  return {
    experiment,
    banditSnapshot,
    serviceHealth: healthResult.health,
    probes: {
      conversionIngest: manualProbe(
        "Live POST route exists, but this page does not auto-probe it to avoid writing conversions during a read-only page load.",
      ),
      pendingById: pendingById.probe,
      userPending: userPending.probe,
    },
  };
}

export async function getDelayedFeedbackDashboardFromCookies(): Promise<DelayedFeedbackDashboardData> {
  const experiments = await getBanditExperimentsFromCookies();
  if (!experiments) {
    return { experiments: [], selectedExperimentId: null, snapshot: null, loadFailed: true };
  }

  const selected = pickDefaultExperiment(experiments);
  if (!selected) {
    return { experiments, selectedExperimentId: null, snapshot: null, loadFailed: false };
  }

  const snapshot = await getDelayedFeedbackSnapshotFromCookies(selected.id);
  return {
    experiments,
    selectedExperimentId: selected.id,
    snapshot,
    loadFailed: false,
  };
}
