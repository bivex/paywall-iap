"use client";

import { type FormEvent, useEffect, useMemo, useState, useTransition } from "react";

import Link from "next/link";

import { Activity, Brain, CalendarClock, FlaskConical, RefreshCw, Settings2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import {
  completeExperimentAction,
  lockExperimentAction,
  pauseExperimentAction,
  repairExperimentAction,
  resumeExperimentAction,
  unlockExperimentAction,
  updateExperimentAction,
} from "@/actions/experiments";
import { PricingTierManager } from "@/components/pricing/pricing-tier-manager";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import type {
  ExperimentStudioDashboardData,
  ExperimentStudioSnapshot,
  StudioEndpointProbe,
} from "@/lib/experiment-studio";
import {
  type ExperimentAlgorithm,
  type ExperimentRepairSummary,
  type ExperimentStatus,
  type ExperimentSummary,
  formatExperimentLifecycleCode,
  getExperimentLifecycleReason,
  getExperimentLifecycleReasonKey,
  getExperimentLifecycleSourceKey,
} from "@/lib/experiments";
import type { PricingTier } from "@/lib/pricing-tiers";

async function fetchStudioJson<T>(url: string): Promise<T> {
  const res = await fetch(url, { cache: "no-store" });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(
      (body as { message?: string; error?: string }).message ??
        (body as { error?: string }).error ??
        `HTTP ${res.status}`,
    );
  }

  return ((body as { data?: T }).data ?? body) as T;
}

function _formatAlgorithm(value: ExperimentAlgorithm | null) {
  if (!value) return "—";
  return value.replaceAll("_", " ");
}

function statusClass(status: ExperimentStatus) {
  switch (status) {
    case "running":
      return "bg-green-100 text-green-800";
    case "draft":
      return "bg-yellow-100 text-yellow-800";
    case "paused":
      return "bg-orange-100 text-orange-800";
    case "completed":
      return "bg-gray-100 text-gray-700";
  }
}

function endpointClass(ok: boolean) {
  return ok ? "bg-green-100 text-green-800" : "bg-yellow-100 text-yellow-800";
}

function formatDate(value: string | null) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("en-US", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

function formatRevenue(value: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

function formatPercent(value: number | null | undefined, digits = 1) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return `${(value * 100).toFixed(digits)}%`;
}

function formatPercentNumber(value: number | null | undefined, digits = 1) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return `${value.toFixed(digits)}%`;
}

function toDatetimeLocalValue(value: string | null | undefined) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

function hasActiveTimedLock(value: string | null | undefined) {
  if (!value) return false;
  const date = new Date(value);
  return !Number.isNaN(date.getTime()) && date.getTime() > Date.now();
}

function formatLifecycleReason(t: ReturnType<typeof useTranslations>, experiment: ExperimentSummary) {
  const reason = getExperimentLifecycleReason(experiment.latest_lifecycle_audit);
  const reasonKey = getExperimentLifecycleReasonKey(reason);
  return reasonKey ? t(`lifecycle.${reasonKey}`) : formatExperimentLifecycleCode(reason);
}

function formatLifecycleSource(t: ReturnType<typeof useTranslations>, experiment: ExperimentSummary) {
  const sourceKey = getExperimentLifecycleSourceKey(experiment.latest_lifecycle_audit?.source);
  return sourceKey
    ? t(`lifecycle.${sourceKey}`)
    : formatExperimentLifecycleCode(experiment.latest_lifecycle_audit?.source);
}

function formatLifecycleHistoryReason(t: ReturnType<typeof useTranslations>, reason: string | null | undefined) {
  const reasonKey = getExperimentLifecycleReasonKey(reason);
  return reasonKey ? t(`lifecycle.${reasonKey}`) : formatExperimentLifecycleCode(reason);
}

function formatLifecycleHistorySource(t: ReturnType<typeof useTranslations>, source: string | null | undefined) {
  const sourceKey = getExperimentLifecycleSourceKey(source);
  return sourceKey ? t(`lifecycle.${sourceKey}`) : formatExperimentLifecycleCode(source);
}

function weightShare(arms: ExperimentSummary["arms"], trafficWeight: number) {
  const total = arms.reduce((sum, arm) => sum + Math.max(arm.traffic_weight, 0), 0);
  return total > 0 ? trafficWeight / total : 0;
}

type DraftMetadataFormValues = {
  name: string;
  description: string;
  algorithm_type: ExperimentAlgorithm;
  is_bandit: boolean;
  min_sample_size: string;
  confidence_threshold_percent: string;
  start_at: string;
  end_at: string;
};

function toDateTimeLocalValue(value: string | null) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

function buildDraftMetadataForm(experiment: ExperimentSummary): DraftMetadataFormValues {
  return {
    name: experiment.name,
    description: experiment.description,
    algorithm_type: experiment.algorithm_type ?? "thompson_sampling",
    is_bandit: experiment.is_bandit,
    min_sample_size: experiment.min_sample_size.toString(),
    confidence_threshold_percent: experiment.confidence_threshold_percent.toString(),
    start_at: toDateTimeLocalValue(experiment.start_at),
    end_at: toDateTimeLocalValue(experiment.end_at),
  };
}

function leadingArm(snapshot: ExperimentStudioSnapshot | null) {
  const probabilities = snapshot?.banditSnapshot?.statistics?.win_probabilities;
  const experiment = snapshot?.experiment;
  if (!probabilities || !experiment) return null;

  return (
    [...experiment.arms]
      .map((arm) => ({ arm, probability: probabilities[arm.id] ?? -1 }))
      .sort((left, right) => right.probability - left.probability)[0] ?? null
  );
}

export function StudioPageClient({
  initialExperiments,
  initialSelectedExperimentId = null,
  initialSnapshot = null,
  initialPricingTiers = [],
  pricingLoadFailed: initialPricingLoadFailed = false,
  loadFailed: initialLoadFailed = false,
}: {
  initialExperiments?: ExperimentSummary[];
  initialSelectedExperimentId?: string | null;
  initialSnapshot?: ExperimentStudioSnapshot | null;
  initialPricingTiers?: PricingTier[];
  pricingLoadFailed?: boolean;
  loadFailed?: boolean;
}) {
  const hasInitialPayload = initialExperiments !== undefined;
  const t = useTranslations("experimentStudio");
  const [hasHydrated, setHasHydrated] = useState(false);
  const [experiments, setExperiments] = useState<ExperimentSummary[]>(initialExperiments ?? []);
  const [selectedId, setSelectedId] = useState(initialSelectedExperimentId ?? "");
  const [snapshot, setSnapshot] = useState<ExperimentStudioSnapshot | null>(initialSnapshot ?? null);
  const [pricingTiers, setPricingTiers] = useState<PricingTier[]>(initialPricingTiers);
  const [pricingLoadFailed, setPricingLoadFailed] = useState(initialPricingLoadFailed);
  const [loadFailed, setLoadFailed] = useState(initialLoadFailed);
  const [isBootstrapping, setIsBootstrapping] = useState(!hasInitialPayload);
  const [isPending, startTransition] = useTransition();
  const [pendingLifecycleAction, setPendingLifecycleAction] = useState<
    "pause" | "resume" | "complete" | "lock" | "unlock" | "repair" | null
  >(null);
  const [pendingSave, setPendingSave] = useState(false);
  const [lockReason, setLockReason] = useState("");
  const [lockUntil, setLockUntil] = useState("");
  const [repairSummary, setRepairSummary] = useState<ExperimentRepairSummary | null>(null);

  useEffect(() => {
    setHasHydrated(true);
  }, []);

  useEffect(() => {
    if (!isBootstrapping) return;

    startTransition(async () => {
      try {
        const data = await fetchStudioJson<ExperimentStudioDashboardData>("/api/admin/studio/dashboard");
        setExperiments(data.experiments);
        setSelectedId(data.selectedExperimentId ?? "");
        setSnapshot(data.snapshot);
        setPricingTiers(data.pricingTiers ?? []);
        setPricingLoadFailed(data.pricingLoadFailed);
        setLoadFailed(data.loadFailed);
      } catch {
        setLoadFailed(true);
        setPricingLoadFailed(true);
      } finally {
        setIsBootstrapping(false);
      }
    });
  }, [isBootstrapping]);

  const selectedExperiment = useMemo(
    () => experiments.find((experiment) => experiment.id === selectedId) ?? snapshot?.experiment ?? null,
    [experiments, selectedId, snapshot],
  );
  const currentLeadingArm = useMemo(() => leadingArm(snapshot), [snapshot]);
  const draftMetadataBaseline = useMemo(
    () => (selectedExperiment ? buildDraftMetadataForm(selectedExperiment) : null),
    [selectedExperiment],
  );
  const [draftMetadata, setDraftMetadata] = useState<DraftMetadataFormValues | null>(draftMetadataBaseline);

  useEffect(() => {
    setDraftMetadata(draftMetadataBaseline);
  }, [draftMetadataBaseline]);

  useEffect(() => {
    setLockReason(selectedExperiment?.automation_policy?.lock_reason ?? "");
    setLockUntil(toDatetimeLocalValue(selectedExperiment?.automation_policy?.locked_until));
  }, [selectedExperiment?.automation_policy?.lock_reason, selectedExperiment?.automation_policy?.locked_until]);

  const isDraftDirty =
    draftMetadata !== null &&
    draftMetadataBaseline !== null &&
    JSON.stringify(draftMetadata) !== JSON.stringify(draftMetadataBaseline);

  function syncExperiment(updatedExperiment: ExperimentSummary) {
    setExperiments((current) => current.map((item) => (item.id === updatedExperiment.id ? updatedExperiment : item)));
    setSnapshot((current) =>
      current?.experiment.id === updatedExperiment.id ? { ...current, experiment: updatedExperiment } : current,
    );
  }

  const lifecycleActions: Array<{ key: "pause" | "resume" | "complete"; label: string }> =
    selectedExperiment?.status === "draft"
      ? [{ key: "resume", label: t("actions.start") }]
      : selectedExperiment?.status === "running"
        ? [
            { key: "pause", label: t("actions.pause") },
            { key: "complete", label: t("actions.complete") },
          ]
        : selectedExperiment?.status === "paused"
          ? [
              { key: "resume", label: t("actions.resume") },
              { key: "complete", label: t("actions.complete") },
            ]
          : [];
  const hasManualOverride = Boolean(selectedExperiment?.automation_policy?.manual_override);
  const hasTimedLock = hasActiveTimedLock(selectedExperiment?.automation_policy?.locked_until);
  const schedulerLocked = hasManualOverride || hasTimedLock;

  const refreshSnapshot = (experimentId: string) => {
    setSelectedId(experimentId);
    setRepairSummary(null);
    startTransition(async () => {
      try {
        const data = await fetchStudioJson<ExperimentStudioSnapshot>(
          `/api/admin/studio/snapshot?experimentId=${encodeURIComponent(experimentId)}`,
        );
        setSnapshot(data);
      } catch {
        toast.error(t("feedback.loadFailed"));
      }
    });
  };

  async function runLifecycleAction(action: "pause" | "resume" | "complete") {
    if (!selectedExperiment) return;

    setPendingLifecycleAction(action);
    const result =
      action === "pause"
        ? await pauseExperimentAction(selectedExperiment.id)
        : action === "complete"
          ? await completeExperimentAction(selectedExperiment.id)
          : await resumeExperimentAction(selectedExperiment.id);
    setPendingLifecycleAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.statusFailed"));
      return;
    }

    syncExperiment(result.data);
    toast.success(t("feedback.statusUpdated"));
    refreshSnapshot(result.data.id);
  }

  async function runLockAction() {
    if (!selectedExperiment) return;

    let lockedUntilISO: string | null = null;
    if (lockUntil) {
      const parsed = new Date(lockUntil);
      if (Number.isNaN(parsed.getTime())) {
        toast.error(t("feedback.lockFailed"));
        return;
      }
      lockedUntilISO = parsed.toISOString();
    }

    setPendingLifecycleAction("lock");
    const result = await lockExperimentAction(selectedExperiment.id, {
      locked_until: lockedUntilISO,
      reason: lockReason,
    });
    setPendingLifecycleAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.lockFailed"));
      return;
    }

    syncExperiment(result.data);
    toast.success(lockedUntilISO ? t("feedback.lockTimed") : t("feedback.locked"));
    refreshSnapshot(result.data.id);
  }

  async function runUnlockAction() {
    if (!selectedExperiment) return;

    setPendingLifecycleAction("unlock");
    const result = await unlockExperimentAction(selectedExperiment.id);
    setPendingLifecycleAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.unlockFailed"));
      return;
    }

    syncExperiment(result.data);
    toast.success(t("feedback.unlocked"));
    refreshSnapshot(result.data.id);
  }

  async function runRepairAction() {
    if (!selectedExperiment) return;

    setPendingLifecycleAction("repair");
    const result = await repairExperimentAction(selectedExperiment.id);
    setPendingLifecycleAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.repairFailed"));
      return;
    }

    syncExperiment(result.data.experiment);
    setRepairSummary(result.data.summary);
    toast.success(t("feedback.repaired"));
    refreshSnapshot(result.data.experiment.id);
  }

  async function saveDraftMetadata(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedExperiment || selectedExperiment.status !== "draft" || !draftMetadata) return;

    setPendingSave(true);
    const result = await updateExperimentAction(selectedExperiment.id, {
      name: draftMetadata.name.trim(),
      description: draftMetadata.description.trim(),
      algorithm_type: draftMetadata.algorithm_type,
      is_bandit: draftMetadata.is_bandit,
      min_sample_size: Number(draftMetadata.min_sample_size),
      confidence_threshold_percent: Number(draftMetadata.confidence_threshold_percent),
      start_at: draftMetadata.start_at ? new Date(draftMetadata.start_at).toISOString() : null,
      end_at: draftMetadata.end_at ? new Date(draftMetadata.end_at).toISOString() : null,
    });
    setPendingSave(false);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.saveFailed"));
      return;
    }

    syncExperiment(result.data);
    toast.success(t("feedback.saved"));
    refreshSnapshot(result.data.id);
  }

  const capabilityRows: Array<{ key: string; probe: StudioEndpointProbe }> = snapshot
    ? [
        { key: "statistics", probe: snapshot.endpoints.statistics },
        { key: "metrics", probe: snapshot.endpoints.metrics },
        { key: "objectives", probe: snapshot.endpoints.objectives },
        { key: "windowInfo", probe: snapshot.endpoints.windowInfo },
      ]
    : [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-semibold text-2xl tracking-tight">{t("title")}</h1>
          <p className="mt-0.5 text-muted-foreground text-sm">{t("subtitle")}</p>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button asChild variant="outline" size="sm">
            <Link href="/dashboard/experiments">
              <FlaskConical className="size-4" />
              {t("actions.openExperiments")}
            </Link>
          </Button>
          <Button asChild variant="outline" size="sm">
            <Link href="/dashboard/experiments/bandit">
              <Brain className="size-4" />
              {t("actions.openBandit")}
            </Link>
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => selectedId && refreshSnapshot(selectedId)}
            disabled={!selectedId || isPending}
          >
            <RefreshCw className={`size-4 ${isPending ? "animate-spin" : ""}`} />
            {t("actions.refresh")}
          </Button>
        </div>
      </div>

      {loadFailed ? (
        <Card className="border-destructive/40">
          <CardContent className="pt-6 text-destructive text-sm">{t("states.loadFailed")}</CardContent>
        </Card>
      ) : null}

      {isBootstrapping ? (
        <Card>
          <CardContent className="pt-6 text-muted-foreground text-sm">{t("states.loading")}</CardContent>
        </Card>
      ) : experiments.length === 0 ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t("states.emptyTitle")}</CardTitle>
            <CardDescription>{t("states.emptyBody")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild size="sm">
              <Link href="/dashboard/experiments">
                <Settings2 className="size-4" />
                {t("actions.openExperiments")}
              </Link>
            </Button>
          </CardContent>
        </Card>
      ) : (
        <>
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("selector.title")}</CardTitle>
              <CardDescription>{t("selector.description")}</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-wrap items-center gap-3">
              <Select value={selectedId} onValueChange={refreshSnapshot}>
                <SelectTrigger className="w-full sm:w-[420px]">
                  <SelectValue placeholder={t("selector.placeholder")} />
                </SelectTrigger>
                <SelectContent>
                  {experiments.map((experiment) => (
                    <SelectItem key={experiment.id} value={experiment.id}>
                      {experiment.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {selectedExperiment ? (
                <div className="flex flex-wrap items-center gap-2">
                  <Badge className={statusClass(selectedExperiment.status)}>
                    {t(`status.${selectedExperiment.status}`)}
                  </Badge>
                  <Badge variant="outline">{selectedExperiment.is_bandit ? t("mode.bandit") : t("mode.classic")}</Badge>
                </div>
              ) : null}
            </CardContent>
          </Card>

          {selectedExperiment ? (
            <>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-5">
                {[
                  {
                    label: t("summary.mode"),
                    value: selectedExperiment.is_bandit ? t("mode.bandit") : t("mode.classic"),
                  },
                  { label: t("summary.arms"), value: selectedExperiment.arm_count.toLocaleString("en-US") },
                  { label: t("summary.samples"), value: selectedExperiment.total_samples.toLocaleString("en-US") },
                  { label: t("summary.revenue"), value: formatRevenue(selectedExperiment.total_revenue) },
                  {
                    label: t("summary.runtime"),
                    value: snapshot?.banditHealth
                      ? `${snapshot.banditHealth.service}: ${snapshot.banditHealth.status}`
                      : t("summary.unavailable"),
                  },
                ].map((item) => (
                  <Card key={item.label}>
                    <CardContent className="pt-6">
                      <p className="font-semibold text-muted-foreground text-xs uppercase tracking-widest">
                        {item.label}
                      </p>
                      <p className="mt-2 break-all font-bold text-lg">{item.value}</p>
                    </CardContent>
                  </Card>
                ))}
              </div>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("editor.title")}</CardTitle>
                  <CardDescription>
                    {selectedExperiment.status === "draft" ? t("editor.description") : t("editor.lockedDescription")}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  {selectedExperiment.status === "draft" && draftMetadata ? (
                    <form className="space-y-4" onSubmit={(event) => void saveDraftMetadata(event)}>
                      <div className="grid gap-4 md:grid-cols-2">
                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("editor.name")}</p>
                          <Input
                            placeholder={t("editor.namePlaceholder")}
                            value={draftMetadata.name}
                            onChange={(event) =>
                              setDraftMetadata((current) =>
                                current ? { ...current, name: event.target.value } : current,
                              )
                            }
                          />
                        </div>

                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("editor.algorithm")}</p>
                          <Select
                            value={draftMetadata.algorithm_type}
                            onValueChange={(value) =>
                              setDraftMetadata((current) =>
                                current ? { ...current, algorithm_type: value as ExperimentAlgorithm } : current,
                              )
                            }
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="thompson_sampling">Thompson Sampling</SelectItem>
                              <SelectItem value="ucb">UCB</SelectItem>
                              <SelectItem value="epsilon_greedy">Epsilon Greedy</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>

                        <div className="space-y-1 md:col-span-2">
                          <p className="font-medium text-xs">{t("editor.descriptionLabel")}</p>
                          <Textarea
                            placeholder={t("editor.descriptionPlaceholder")}
                            rows={3}
                            value={draftMetadata.description}
                            onChange={(event) =>
                              setDraftMetadata((current) =>
                                current ? { ...current, description: event.target.value } : current,
                              )
                            }
                          />
                        </div>

                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("editor.minSampleSize")}</p>
                          <Input
                            inputMode="numeric"
                            placeholder="100"
                            value={draftMetadata.min_sample_size}
                            onChange={(event) =>
                              setDraftMetadata((current) =>
                                current ? { ...current, min_sample_size: event.target.value } : current,
                              )
                            }
                          />
                        </div>

                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("editor.confidenceThreshold")}</p>
                          <Input
                            inputMode="decimal"
                            placeholder="95"
                            value={draftMetadata.confidence_threshold_percent}
                            onChange={(event) =>
                              setDraftMetadata((current) =>
                                current ? { ...current, confidence_threshold_percent: event.target.value } : current,
                              )
                            }
                          />
                        </div>

                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("editor.startAt")}</p>
                          <Input
                            type="datetime-local"
                            value={hasHydrated ? draftMetadata.start_at : ""}
                            onChange={(event) =>
                              setDraftMetadata((current) =>
                                current ? { ...current, start_at: event.target.value } : current,
                              )
                            }
                          />
                        </div>

                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("editor.endAt")}</p>
                          <Input
                            type="datetime-local"
                            value={hasHydrated ? draftMetadata.end_at : ""}
                            onChange={(event) =>
                              setDraftMetadata((current) =>
                                current ? { ...current, end_at: event.target.value } : current,
                              )
                            }
                          />
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <Switch
                          checked={draftMetadata.is_bandit}
                          onCheckedChange={(checked) =>
                            setDraftMetadata((current) => (current ? { ...current, is_bandit: checked } : current))
                          }
                        />
                        <span className="text-sm">{t("editor.isBandit")}</span>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        <Button type="submit" size="sm" disabled={pendingSave || isPending || !isDraftDirty}>
                          {pendingSave ? t("feedback.saving") : t("actions.save")}
                        </Button>
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          disabled={pendingSave || !isDraftDirty}
                          onClick={() => setDraftMetadata(draftMetadataBaseline)}
                        >
                          {t("actions.reset")}
                        </Button>
                      </div>
                    </form>
                  ) : (
                    <div className="rounded-md border border-dashed p-4 text-muted-foreground text-sm">
                      {t("editor.lockedBody")}
                    </div>
                  )}
                </CardContent>
              </Card>

              <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.1fr_1fr]">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm">{t("lifecycle.title")}</CardTitle>
                    <CardDescription>{t("lifecycle.description")}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    <div className="grid gap-3 md:grid-cols-2">
                      {[
                        {
                          label: t("lifecycle.created"),
                          value: hasHydrated ? formatDate(selectedExperiment.created_at) : "—",
                        },
                        {
                          label: t("lifecycle.updated"),
                          value: hasHydrated ? formatDate(selectedExperiment.updated_at) : "—",
                        },
                        {
                          label: t("lifecycle.start"),
                          value: hasHydrated ? formatDate(selectedExperiment.start_at) : "—",
                        },
                        { label: t("lifecycle.end"), value: hasHydrated ? formatDate(selectedExperiment.end_at) : "—" },
                      ].map((item) => (
                        <div key={item.label} className="rounded-md border p-4">
                          <p className="font-medium text-sm">{item.label}</p>
                          <p className="mt-1 text-muted-foreground text-sm">{item.value}</p>
                        </div>
                      ))}
                    </div>

                    <div className="rounded-md border border-dashed p-4">
                      <p className="font-medium text-sm">{t("lifecycle.controlsTitle")}</p>
                      <p className="mt-1 text-muted-foreground text-xs">{t("lifecycle.controlsDescription")}</p>

                      <div className="mt-3 flex flex-wrap gap-2">
                        {lifecycleActions.length === 0 ? (
                          <span className="text-muted-foreground text-xs">{t("lifecycle.noActions")}</span>
                        ) : (
                          lifecycleActions.map((action) => (
                            <Button
                              key={action.key}
                              variant="outline"
                              size="sm"
                              disabled={isPending || pendingLifecycleAction !== null}
                              onClick={() => void runLifecycleAction(action.key)}
                            >
                              {pendingLifecycleAction === action.key ? t("feedback.statusUpdating") : action.label}
                            </Button>
                          ))
                        )}
                      </div>
                    </div>

                    <div className="rounded-md border p-4">
                      <div className="flex flex-wrap items-start justify-between gap-3">
                        <div>
                          <p className="font-medium text-sm">{t("lifecycle.lockTitle")}</p>
                          <p className="mt-1 text-muted-foreground text-xs">{t("lifecycle.lockDescription")}</p>
                        </div>
                        <Badge variant={schedulerLocked ? "default" : "outline"}>
                          {schedulerLocked ? t("lifecycle.lockActive") : t("lifecycle.lockInactive")}
                        </Badge>
                      </div>

                      <div className="mt-3 flex flex-wrap gap-2">
                        {hasManualOverride ? (
                          <Badge variant="outline">{t("lifecycle.manualOverrideActive")}</Badge>
                        ) : null}
                        {hasTimedLock ? <Badge variant="outline">{t("lifecycle.timedLockActive")}</Badge> : null}
                        {!hasTimedLock && selectedExperiment?.automation_policy?.locked_until ? (
                          <Badge variant="outline">{t("lifecycle.timedLockExpired")}</Badge>
                        ) : null}
                      </div>

                      <div className="mt-4 grid gap-3 md:grid-cols-2">
                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("lifecycle.lockUntilLabel")}</p>
                          <Input
                            type="datetime-local"
                            value={lockUntil}
                            onChange={(event) => setLockUntil(event.target.value)}
                            disabled={isPending || pendingLifecycleAction !== null}
                          />
                          <p className="text-[11px] text-muted-foreground">{t("lifecycle.lockUntilHint")}</p>
                        </div>
                        <div className="space-y-1">
                          <p className="font-medium text-xs">{t("lifecycle.lockReasonLabel")}</p>
                          <Input
                            value={lockReason}
                            onChange={(event) => setLockReason(event.target.value)}
                            placeholder={t("lifecycle.lockReasonPlaceholder")}
                            disabled={isPending || pendingLifecycleAction !== null}
                          />
                        </div>
                      </div>

                      <div className="mt-4 space-y-1 text-muted-foreground text-xs">
                        <p>
                          {t("lifecycle.lockedUntilValue")}:{" "}
                          {formatDate(selectedExperiment?.automation_policy?.locked_until ?? null)}
                        </p>
                        <p>
                          {t("lifecycle.lockedByValue")}: {selectedExperiment?.automation_policy?.locked_by ?? "—"}
                        </p>
                        <p>
                          {t("lifecycle.lockReasonValue")}: {selectedExperiment?.automation_policy?.lock_reason || "—"}
                        </p>
                      </div>

                      <div className="mt-4 flex flex-wrap gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={isPending || pendingLifecycleAction !== null}
                          onClick={() => void runLockAction()}
                        >
                          {pendingLifecycleAction === "lock" ? t("feedback.locking") : t("actions.lock")}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={isPending || pendingLifecycleAction !== null || !schedulerLocked}
                          onClick={() => void runUnlockAction()}
                        >
                          {pendingLifecycleAction === "unlock" ? t("feedback.unlocking") : t("actions.unlock")}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          disabled={isPending || pendingLifecycleAction !== null}
                          onClick={() => void runRepairAction()}
                        >
                          {pendingLifecycleAction === "repair" ? t("feedback.repairing") : t("actions.repair")}
                        </Button>
                      </div>

                      {repairSummary ? (
                        <div className="mt-4 rounded-md border border-dashed p-3 text-xs">
                          <p className="font-medium text-sm">{t("lifecycle.repairSummaryTitle")}</p>
                          <div className="mt-2 space-y-1 text-muted-foreground">
                            <p>
                              {t("lifecycle.repairAssignments")}: {repairSummary.assignment_snapshot.active}/
                              {repairSummary.assignment_snapshot.total}
                            </p>
                            <p>
                              {t("lifecycle.repairArmStats")}: {repairSummary.missing_arm_stats_inserted}
                            </p>
                            <p>
                              {t("lifecycle.repairPendingRewards")}: {repairSummary.pending_rewards_processed}/
                              {repairSummary.expired_pending_rewards}
                            </p>
                            <p>
                              {t("lifecycle.repairWinnerConfidence")}:{" "}
                              {formatPercentNumber(repairSummary.winner_confidence_percent)}
                            </p>
                          </div>
                        </div>
                      ) : null}
                    </div>

                    <div className="rounded-md border p-4">
                      <p className="font-medium text-sm">{t("lifecycle.lastActionTitle")}</p>
                      {selectedExperiment.latest_lifecycle_audit ? (
                        <div className="mt-3 space-y-2 text-sm">
                          <div className="flex flex-wrap gap-2">
                            <Badge variant="outline">{formatLifecycleSource(t, selectedExperiment)}</Badge>
                            <Badge variant="outline">{formatLifecycleReason(t, selectedExperiment)}</Badge>
                          </div>
                          <p className="text-muted-foreground text-xs">
                            {t("lifecycle.lastActionTransition")}:{" "}
                            {t(`status.${selectedExperiment.latest_lifecycle_audit.from_status}`)} →{" "}
                            {t(`status.${selectedExperiment.latest_lifecycle_audit.to_status}`)}
                          </p>
                          <p className="text-muted-foreground text-xs">
                            {t("lifecycle.lastActionAt")}:{" "}
                            {hasHydrated ? formatDate(selectedExperiment.latest_lifecycle_audit.created_at) : "—"}
                          </p>
                        </div>
                      ) : (
                        <p className="mt-2 text-muted-foreground text-xs">{t("lifecycle.lastActionEmpty")}</p>
                      )}
                    </div>

                    <div className="rounded-md border p-4">
                      <p className="font-medium text-sm">{t("lifecycle.historyTitle")}</p>
                      <p className="mt-1 text-muted-foreground text-xs">{t("lifecycle.historyDescription")}</p>
                      {snapshot?.lifecycleHistory.length ? (
                        <div className="mt-3 space-y-2">
                          {snapshot.lifecycleHistory.map((entry, index) => {
                            const reason = typeof entry.details?.reason === "string" ? entry.details.reason : null;
                            return (
                              <div
                                key={`${entry.created_at}-${entry.action}-${index}`}
                                className="rounded-md border border-dashed p-3"
                              >
                                <div className="flex flex-wrap gap-2">
                                  <Badge variant="outline">{formatLifecycleHistorySource(t, entry.source)}</Badge>
                                  <Badge variant="outline">{formatLifecycleHistoryReason(t, reason)}</Badge>
                                </div>
                                <p className="mt-2 text-muted-foreground text-xs">
                                  {t(`status.${entry.from_status}`)} → {t(`status.${entry.to_status}`)}
                                </p>
                                <p className="mt-1 text-muted-foreground text-xs">
                                  {hasHydrated ? formatDate(entry.created_at) : "—"}
                                </p>
                              </div>
                            );
                          })}
                        </div>
                      ) : (
                        <p className="mt-2 text-muted-foreground text-xs">{t("lifecycle.historyEmpty")}</p>
                      )}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm">{t("capabilities.title")}</CardTitle>
                    <CardDescription>{t("capabilities.description")}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {capabilityRows.map((item) => (
                      <div key={item.key} className="rounded-md border p-3">
                        <div className="flex items-center justify-between gap-3">
                          <p className="font-medium text-sm">{t(`capabilities.${item.key}`)}</p>
                          <Badge className={endpointClass(item.probe.ok)}>
                            {item.probe.ok ? t("capabilities.available") : t("capabilities.unavailable")}
                          </Badge>
                        </div>
                        <p className="mt-1 text-muted-foreground text-xs">
                          {item.probe.status ? `HTTP ${item.probe.status} · ` : ""}
                          {item.probe.message}
                        </p>
                      </div>
                    ))}
                  </CardContent>
                </Card>
              </div>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("arms.title")}</CardTitle>
                  <CardDescription>{t("arms.description")}</CardDescription>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t("table.arm")}</TableHead>
                        <TableHead>{t("table.type")}</TableHead>
                        <TableHead>{t("table.weight")}</TableHead>
                        <TableHead>{t("table.samples")}</TableHead>
                        <TableHead>{t("table.conversions")}</TableHead>
                        <TableHead>{t("table.winProbability")}</TableHead>
                        <TableHead>{t("table.revenue")}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {selectedExperiment.arms.map((arm) => {
                        const share = weightShare(selectedExperiment.arms, arm.traffic_weight);
                        const winProbability =
                          snapshot?.banditSnapshot?.statistics?.win_probabilities?.[arm.id] ?? null;
                        return (
                          <TableRow key={arm.id}>
                            <TableCell>
                              <div>
                                <p className="font-medium">{arm.name}</p>
                                <p className="max-w-sm text-muted-foreground text-xs">{arm.description || "—"}</p>
                              </div>
                            </TableCell>
                            <TableCell>{arm.is_control ? t("table.control") : t("table.variant")}</TableCell>
                            <TableCell className="min-w-36">
                              <div className="space-y-1">
                                <Progress value={share * 100} className="h-2" />
                                <p className="text-muted-foreground text-xs">
                                  {arm.traffic_weight.toFixed(2)} · {formatPercent(share, 0)}
                                </p>
                              </div>
                            </TableCell>
                            <TableCell>{arm.samples.toLocaleString("en-US")}</TableCell>
                            <TableCell>{arm.conversions.toLocaleString("en-US")}</TableCell>
                            <TableCell>{formatPercent(winProbability)}</TableCell>
                            <TableCell>{formatRevenue(arm.revenue)}</TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>

              <div className="space-y-3">
                <div>
                  <h2 className="font-semibold text-lg tracking-tight">{t("pricing.title")}</h2>
                  <p className="mt-1 text-muted-foreground text-sm">{t("pricing.description")}</p>
                </div>

                <div className="rounded-md border border-dashed p-4 text-muted-foreground text-xs">
                  <p className="font-medium text-foreground text-sm">{t("pricing.noteTitle")}</p>
                  <p className="mt-1">{t("pricing.noteBody")}</p>
                </div>

                <PricingTierManager embedded initialTiers={pricingTiers} loadFailed={pricingLoadFailed} />
              </div>

              <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.1fr_1fr]">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm">{t("runtime.title")}</CardTitle>
                    <CardDescription>{t("runtime.description")}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3 text-sm">
                    <div className="flex items-center justify-between rounded-md border p-3">
                      <span>{t("runtime.leadingArm")}</span>
                      <span className="font-mono">
                        {currentLeadingArm
                          ? `${currentLeadingArm.arm.name} · ${formatPercent(currentLeadingArm.probability)}`
                          : "—"}
                      </span>
                    </div>
                    <div className="flex items-center justify-between rounded-md border p-3">
                      <span>{t("runtime.balanceIndex")}</span>
                      <span className="font-mono">
                        {snapshot?.banditSnapshot?.metrics
                          ? formatPercent(snapshot.banditSnapshot.metrics.balance_index)
                          : t("summary.unavailable")}
                      </span>
                    </div>
                    <div className="flex items-center justify-between rounded-md border p-3">
                      <span>{t("runtime.pendingRewards")}</span>
                      <span className="font-mono">{snapshot?.banditSnapshot?.metrics?.pending_rewards ?? 0}</span>
                    </div>
                    <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                      <Activity className="mb-2 size-4" />
                      {selectedExperiment.is_bandit ? t("runtime.banditNote") : t("runtime.classicNote")}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-sm">{t("notes.title")}</CardTitle>
                    <CardDescription>{t("notes.description")}</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-3 text-sm">
                    <div className="rounded-md border p-3">
                      <p className="font-medium">{t("notes.builderTitle")}</p>
                      <p className="mt-1 text-muted-foreground text-xs">{t("notes.builderBody")}</p>
                    </div>
                    <div className="rounded-md border p-3">
                      <p className="font-medium">{t("notes.banditTitle")}</p>
                      <p className="mt-1 text-muted-foreground text-xs">{t("notes.banditBody")}</p>
                    </div>
                    <div className="rounded-md border p-3">
                      <p className="font-medium">{t("notes.pricingTitle")}</p>
                      <p className="mt-1 text-muted-foreground text-xs">{t("notes.pricingBody")}</p>
                    </div>
                    <div className="rounded-md border p-3">
                      <p className="font-medium">{t("notes.advancedTitle")}</p>
                      <p className="mt-1 text-muted-foreground text-xs">{t("notes.advancedBody")}</p>
                    </div>
                    <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                      <CalendarClock className="mb-2 size-4" />
                      {t("notes.readOnly")}
                    </div>
                  </CardContent>
                </Card>
              </div>
            </>
          ) : null}
        </>
      )}
    </div>
  );
}
