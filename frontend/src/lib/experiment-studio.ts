import type { BanditSnapshot } from "@/lib/bandit";
import type { ExperimentSummary } from "@/lib/experiments";

export interface StudioEndpointProbe {
  ok: boolean;
  status: number | null;
  message: string;
}

export interface StudioBanditHealth {
  status: string;
  service: string;
}

export interface ExperimentStudioSnapshot {
  experiment: ExperimentSummary;
  banditHealth: StudioBanditHealth | null;
  banditSnapshot: BanditSnapshot | null;
  endpoints: {
    statistics: StudioEndpointProbe;
    metrics: StudioEndpointProbe;
    objectives: StudioEndpointProbe;
    windowInfo: StudioEndpointProbe;
  };
}

export interface ExperimentStudioDashboardData {
  experiments: ExperimentSummary[];
  selectedExperimentId: string | null;
  snapshot: ExperimentStudioSnapshot | null;
  loadFailed: boolean;
}
