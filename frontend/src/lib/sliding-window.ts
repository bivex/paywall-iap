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

export interface SlidingWindowRewardEvent {
  UserID: string;
  ArmID: string;
  RewardValue: number;
  Currency: string;
  Timestamp: string;
  ConversionDelay: number | null;
  Metadata: Record<string, unknown> | null;
}

export interface SlidingWindowEventsExport {
  experiment_id: string;
  events: Record<string, SlidingWindowRewardEvent[]>;
  limit: number;
}
