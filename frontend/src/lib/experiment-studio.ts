import type { BanditSnapshot } from "@/lib/bandit";
import type { ExperimentLifecycleAudit, ExperimentSummary } from "@/lib/experiments";
import type { PricingTier } from "@/lib/pricing-tiers";

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
  lifecycleHistory: ExperimentLifecycleAudit[];
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
  pricingTiers: PricingTier[];
  pricingLoadFailed: boolean;
  loadFailed: boolean;
}
