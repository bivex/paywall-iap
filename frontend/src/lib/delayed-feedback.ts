import type { BanditSnapshot } from "@/lib/bandit";
import type { ExperimentSummary } from "@/lib/experiments";

export interface DelayedEndpointProbe {
  state: "available" | "manual" | "unavailable";
  status: number | null;
  message: string;
}

export interface DelayedFeedbackServiceHealth {
  service: string;
  status: string;
}

export interface DelayedFeedbackSnapshot {
  experiment: ExperimentSummary;
  banditSnapshot: BanditSnapshot | null;
  serviceHealth: DelayedFeedbackServiceHealth | null;
  probes: {
    conversionIngest: DelayedEndpointProbe;
    pendingById: DelayedEndpointProbe;
    userPending: DelayedEndpointProbe;
  };
}

export interface DelayedFeedbackDashboardData {
  experiments: ExperimentSummary[];
  selectedExperimentId: string | null;
  snapshot: DelayedFeedbackSnapshot | null;
  loadFailed: boolean;
}

export interface DelayedConversionPayload {
  transactionId: string;
  userId: string;
  conversionValue: number;
  currency: string;
}

export interface DelayedConversionResult {
  ok: boolean;
  status: number;
  message: string;
  transactionId?: string;
}
