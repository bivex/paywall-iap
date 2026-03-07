export type ExperimentStatus = "draft" | "running" | "paused" | "completed";
export type ExperimentAlgorithm = "thompson_sampling" | "ucb" | "epsilon_greedy";

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
