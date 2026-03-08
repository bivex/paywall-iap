export type ExperimentStatus = "draft" | "running" | "paused" | "completed";
export type ExperimentAlgorithm = "thompson_sampling" | "ucb" | "epsilon_greedy";

export interface ExperimentAutomationPolicy {
  readonly enabled: boolean;
  readonly auto_start: boolean;
  readonly auto_complete: boolean;
  readonly complete_on_end_time: boolean;
  readonly complete_on_sample_size: boolean;
  readonly complete_on_confidence: boolean;
  readonly manual_override: boolean;
}

export interface ExperimentLifecycleAudit {
  readonly actor_type: string;
  readonly source: string;
  readonly action: string;
  readonly from_status: ExperimentStatus;
  readonly to_status: ExperimentStatus;
  readonly idempotency_key?: string | null;
  readonly details?: Record<string, unknown> | null;
  readonly created_at: string;
}

export interface ExperimentArm {
  id: string;
  name: string;
  description: string;
  is_control: boolean;
  traffic_weight: number;
  samples: number;
  conversions: number;
  revenue: number;
  avg_reward: number;
}

export interface ExperimentSummary {
  id: string;
  name: string;
  description: string;
  status: ExperimentStatus;
  algorithm_type: ExperimentAlgorithm | null;
  is_bandit: boolean;
  min_sample_size: number;
  confidence_threshold_percent: number;
  winner_confidence_percent: number | null;
  start_at: string | null;
  end_at: string | null;
  automation_policy?: ExperimentAutomationPolicy;
  latest_lifecycle_audit?: ExperimentLifecycleAudit | null;
  created_at: string;
  updated_at: string;
  arm_count: number;
  total_assignments: number;
  active_assignments: number;
  total_samples: number;
  total_conversions: number;
  total_revenue: number;
  arms: ExperimentArm[];
}

export function getExperimentLifecycleReason(audit?: ExperimentLifecycleAudit | null): string | null {
  const reason = audit?.details?.reason;
  return typeof reason === "string" && reason.length > 0 ? reason : null;
}

export function getExperimentLifecycleReasonKey(reason?: string | null): string | null {
  switch (reason) {
    case "auto_start":
      return "reasonAutoStart";
    case "auto_complete_end_time":
      return "reasonAutoCompleteEndTime";
    case "auto_complete_sample_size":
      return "reasonAutoCompleteSampleSize";
    case "auto_complete_confidence":
      return "reasonAutoCompleteConfidence";
    case "manual_paused":
      return "reasonManualPaused";
    case "manual_running":
      return "reasonManualRunning";
    case "manual_completed":
      return "reasonManualCompleted";
    default:
      return null;
  }
}

export function getExperimentLifecycleSourceKey(source?: string | null): string | null {
  switch (source) {
    case "experiment_automation_reconciler":
      return "sourceAutomation";
    case "admin_experiments_api":
      return "sourceAdminApi";
    default:
      return null;
  }
}

export function formatExperimentLifecycleCode(value?: string | null): string {
  if (!value) return "—";
  return value
    .split("_")
    .map((part) => (part.length > 0 ? `${part[0]?.toUpperCase() ?? ""}${part.slice(1)}` : part))
    .join(" ");
}

export interface ExperimentArmInput {
  name: string;
  description: string;
  is_control: boolean;
  traffic_weight: number;
}

export interface ExperimentInput {
  name: string;
  description: string;
  status: ExperimentStatus;
  algorithm_type: ExperimentAlgorithm;
  is_bandit: boolean;
  min_sample_size: number;
  confidence_threshold_percent: number;
  start_at: string | null;
  end_at: string | null;
  arms: ExperimentArmInput[];
}

export interface ExperimentUpdateInput {
  name: string;
  description: string;
  algorithm_type: ExperimentAlgorithm;
  is_bandit: boolean;
  min_sample_size: number;
  confidence_threshold_percent: number;
  start_at: string | null;
  end_at: string | null;
}

export const EMPTY_EXPERIMENT_INPUT: ExperimentInput = {
  name: "",
  description: "",
  status: "draft",
  algorithm_type: "thompson_sampling",
  is_bandit: true,
  min_sample_size: 100,
  confidence_threshold_percent: 95,
  start_at: null,
  end_at: null,
  arms: [
    { name: "Control", description: "", is_control: true, traffic_weight: 1 },
    { name: "Variant A", description: "", is_control: false, traffic_weight: 1 },
  ],
};
