import type { BanditSnapshot } from "@/lib/bandit";
import type { ExperimentArm, ExperimentArmInput, ExperimentLifecycleAudit, ExperimentSummary } from "@/lib/experiments";
import type { PricingTier } from "@/lib/pricing-tiers";

export type DraftExperimentArmValidationCode = "minimumArms" | "singleControl" | "armName" | "trafficWeight";

export interface DraftExperimentArmForm {
  client_id: string;
  id?: string;
  name: string;
  description: string;
  is_control: boolean;
  traffic_weight: string;
  pricing_tier_id: string | null;
}

function nextDraftArmId(prefix: string) {
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
}

function defaultDraftArmName(index: number) {
  return index < 26 ? `Variant ${String.fromCharCode(65 + index)}` : `Variant ${index + 1}`;
}

function toDraftExperimentArm(arm: ExperimentArm): DraftExperimentArmForm {
  return {
    client_id: arm.id || nextDraftArmId("persisted-arm"),
    id: arm.id,
    name: arm.name,
    description: arm.description,
    is_control: arm.is_control,
    traffic_weight: arm.traffic_weight.toString(),
    pricing_tier_id: arm.pricing_tier_id ?? null,
  };
}

function normalizeDraftExperimentArms(arms: DraftExperimentArmForm[]) {
  return arms.map((arm) => ({
    id: arm.id ?? null,
    name: arm.name,
    description: arm.description,
    is_control: arm.is_control,
    traffic_weight: arm.traffic_weight,
    pricing_tier_id: arm.pricing_tier_id ?? null,
  }));
}

export function buildDraftExperimentArms(experiment: ExperimentSummary): DraftExperimentArmForm[] {
  return experiment.arms.map(toDraftExperimentArm);
}

export function createEmptyDraftExperimentArm(index: number): DraftExperimentArmForm {
  return {
    client_id: nextDraftArmId("new-arm"),
    name: defaultDraftArmName(index),
    description: "",
    is_control: false,
    traffic_weight: "1",
    pricing_tier_id: null,
  };
}

export function serializeDraftExperimentArms(arms: DraftExperimentArmForm[]) {
  return JSON.stringify(normalizeDraftExperimentArms(arms));
}

export function toExperimentUpdateArms(arms: DraftExperimentArmForm[]): ExperimentArmInput[] {
  return arms.map((arm) => ({
    id: arm.id,
    name: arm.name.trim(),
    description: arm.description.trim(),
    is_control: arm.is_control,
    traffic_weight: Number(arm.traffic_weight),
    pricing_tier_id: arm.pricing_tier_id ?? null,
  }));
}

export function validateDraftExperimentArms(arms: DraftExperimentArmForm[]): DraftExperimentArmValidationCode | null {
  if (arms.length < 2) return "minimumArms";
  if (arms.filter((arm) => arm.is_control).length !== 1) return "singleControl";
  if (arms.some((arm) => arm.name.trim().length === 0)) return "armName";
  if (arms.some((arm) => !Number.isFinite(Number(arm.traffic_weight)) || Number(arm.traffic_weight) <= 0)) {
    return "trafficWeight";
  }
  return null;
}

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
