import type { BanditSnapshot } from "@/lib/bandit";
import type { ExperimentSummary } from "@/lib/experiments";

export interface SlidingWindowEndpointProbe {
  state: "available" | "manual" | "unavailable";
  status: number | null;
  message: string;
}

export interface SlidingWindowServiceHealth {
  service: string;
  status: string;
}

export interface SlidingWindowSnapshot {
  experiment: ExperimentSummary;
  banditSnapshot: BanditSnapshot | null;
  serviceHealth: SlidingWindowServiceHealth | null;
  probes: {
    windowInfo: SlidingWindowEndpointProbe;
    windowEvents: SlidingWindowEndpointProbe;
    trimWindow: SlidingWindowEndpointProbe;
  };
}

export interface SlidingWindowDashboardData {
  experiments: ExperimentSummary[];
  selectedExperimentId: string | null;
  snapshot: SlidingWindowSnapshot | null;
  loadFailed: boolean;
}

export interface TrimWindowResult {
  ok: boolean;
  status: number;
  message: string;
  experimentId?: string;
}
