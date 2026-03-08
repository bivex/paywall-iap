import type { ExperimentSummary, ExperimentWinnerRecommendationAudit } from "@/lib/experiments";

export interface BanditArmStatistics {
  arm_id: string;
  alpha: number;
  beta: number;
  samples: number;
  conversions: number;
  revenue: number;
  avg_reward: number;
  conversion_rate: number;
}

export interface BanditStatisticsResponse {
  experiment_id: string;
  arms: BanditArmStatistics[];
  win_probabilities?: Record<string, number>;
}

export interface BanditMetrics {
  regret: number;
  exploration_rate: number;
  convergence_gap: number;
  balance_index: number;
  window_utilization: number;
  pending_rewards: number;
}

export interface BanditSnapshot {
  experiment: ExperimentSummary;
  statistics: BanditStatisticsResponse | null;
  metrics: BanditMetrics | null;
  recommendationHistory: ExperimentWinnerRecommendationAudit[];
}
