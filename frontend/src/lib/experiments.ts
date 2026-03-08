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
  readonly locked_until?: string | null;
  readonly locked_by?: string | null;
  readonly lock_reason?: string | null;
}

export interface ExperimentRepairSummary {
  readonly assignment_snapshot: {
    readonly total: number;
    readonly active: number;
  };
  readonly missing_arm_stats_inserted: number;
  readonly pending_rewards_total: number;
  readonly expired_pending_rewards: number;
  readonly pending_rewards_processed: number;
  readonly winner_confidence_percent: number | null;
}

export interface ExperimentRepairResult {
  readonly experiment: ExperimentSummary;
  readonly summary: ExperimentRepairSummary;
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

export interface ExperimentWinnerRecommendation {
  readonly recommended: boolean;
  readonly reason: string;
  readonly winning_arm_id?: string | null;
  readonly winning_arm_name?: string | null;
  readonly confidence_percent?: number | null;
  readonly confidence_threshold_percent: number;
  readonly observed_samples: number;
  readonly min_sample_size: number;
}

export interface ExperimentWinnerRecommendationAudit {
  readonly source: string;
  readonly recommended: boolean;
  readonly reason: string;
  readonly winning_arm_id?: string | null;
  readonly confidence_percent?: number | null;
  readonly confidence_threshold_percent: number;
  readonly observed_samples: number;
  readonly min_sample_size: number;
  readonly details?: Record<string, unknown> | null;
  readonly occurred_at: string;
}

export interface ExperimentArm {
  id: string;
  name: string;
  description: string;
  is_control: boolean;
  traffic_weight: number;
  pricing_tier_id?: string | null;
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
  winner_recommendation?: ExperimentWinnerRecommendation | null;
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

export function getExperimentWinnerRecommendationReasonKey(reason?: string | null): string | null {
  switch (reason) {
    case "draft_experiment":
      return "reasonDraftExperiment";
    case "status_not_eligible":
      return "reasonStatusNotEligible";
    case "insufficient_arms":
      return "reasonInsufficientArms";
    case "insufficient_data":
      return "reasonInsufficientData";
    case "insufficient_sample_size":
      return "reasonInsufficientSampleSize";
    case "confidence_below_threshold":
      return "reasonConfidenceBelowThreshold";
    case "recommend_winner":
      return "reasonRecommendWinner";
    default:
      return null;
  }
}

export function getExperimentWinnerRecommendationSourceKey(source?: string | null): string | null {
  switch (source) {
    case "admin_experiments_api":
      return "sourceAdminApi";
    case "admin_experiments_list":
      return "sourceAdminOverview";
    case "admin_experiments_detail":
      return "sourceAdminDetail";
    case "winner_recommendation_service":
      return "sourceRecommendationService";
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
  id?: string;
  name: string;
  description: string;
  is_control: boolean;
  traffic_weight: number;
  pricing_tier_id?: string | null;
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
  arms?: ExperimentArmInput[];
}

export interface ExperimentAutomationPolicyUpdateInput {
  enabled: boolean;
  auto_start: boolean;
  auto_complete: boolean;
  complete_on_end_time: boolean;
  complete_on_sample_size: boolean;
  complete_on_confidence: boolean;
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
