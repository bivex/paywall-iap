import type { BanditSnapshot } from "@/lib/bandit";
import type { ExperimentSummary } from "@/lib/experiments";

export type ObjectiveType = "conversion" | "ltv" | "revenue" | "hybrid";

export interface ObjectiveEndpointProbe {
  state: "available" | "manual" | "unavailable";
  status: number | null;
  message: string;
}

export interface ObjectiveServiceHealth {
  service: string;
  status: string;
}

export interface ObjectiveScoreEntry {
  objectiveType: string;
  score: number | null;
  alpha: number | null;
  beta: number | null;
  samples: number | null;
  conversions: number | null;
  revenue: number | null;
  avgLtv: number | null;
}

export interface ObjectiveCurrentConfig {
  experimentId: string;
  objectiveType: ObjectiveType;
  weights: Record<string, number>;
}

export type ObjectiveScoresByArm = Record<string, Record<string, ObjectiveScoreEntry>>;

export interface MultiObjectiveSnapshot {
  experiment: ExperimentSummary;
  banditSnapshot: BanditSnapshot | null;
  serviceHealth: ObjectiveServiceHealth | null;
  objectiveScores: ObjectiveScoresByArm | null;
  currentConfig: ObjectiveCurrentConfig | null;
  probes: {
    objectiveScores: ObjectiveEndpointProbe;
    objectiveConfig: ObjectiveEndpointProbe;
  };
}

export interface MultiObjectiveDashboardData {
  experiments: ExperimentSummary[];
  selectedExperimentId: string | null;
  snapshot: MultiObjectiveSnapshot | null;
  loadFailed: boolean;
}

export interface ObjectiveConfigResult {
  ok: boolean;
  status: number;
  message: string;
  objectiveType?: ObjectiveType;
  weights?: Record<string, number>;
}
