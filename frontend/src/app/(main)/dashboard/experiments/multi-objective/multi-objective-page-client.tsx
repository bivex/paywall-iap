"use client";

import { useEffect, useMemo, useState, useTransition } from "react";

import Link from "next/link";

import { Brain, FlaskConical, RefreshCw, Target } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Progress } from "@/components/ui/progress";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { BanditArmStatistics } from "@/lib/bandit";
import type { ExperimentAlgorithm, ExperimentStatus, ExperimentSummary } from "@/lib/experiments";
import type {
  MultiObjectiveDashboardData,
  MultiObjectiveSnapshot,
  ObjectiveConfigResult,
  ObjectiveCurrentConfig,
  ObjectiveEndpointProbe,
  ObjectiveType,
} from "@/lib/multi-objective";

const DEFAULT_OBJECTIVE_TYPE: ObjectiveType = "conversion";
const DEFAULT_OBJECTIVE_WEIGHTS = {
  conversion: "0.5",
  ltv: "0.3",
  revenue: "0.2",
} as const;

const SUPPORTED_OBJECTIVES = [
  { key: "conversion", defaultWeight: 0.5, weightPercent: 50 },
  { key: "ltv", defaultWeight: 0.3, weightPercent: 30 },
  { key: "revenue", defaultWeight: 0.2, weightPercent: 20 },
  { key: "hybrid", defaultWeight: null, weightPercent: null },
] as const;

async function fetchMultiObjectiveJson<T>(url: string): Promise<T> {
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

function formatAlgorithm(value: ExperimentAlgorithm | null) {
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

function probeClass(state: ObjectiveEndpointProbe["state"]) {
  switch (state) {
    case "available":
      return "bg-green-100 text-green-800";
    case "manual":
      return "bg-blue-100 text-blue-800";
    case "unavailable":
      return "bg-yellow-100 text-yellow-800";
  }
}

function formatPercent(value: number | null | undefined, digits = 1) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return `${(value * 100).toFixed(digits)}%`;
}

function formatRatePercent(value: number | null | undefined, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return `${(value * 100).toFixed(digits)}%`;
}

function formatCurrency(value: number | null | undefined) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

function formatScore(value: number | null | undefined) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return value.toFixed(3);
}

function buildWeightInputs(config: ObjectiveCurrentConfig | null | undefined) {
  return {
    conversion: String(config?.weights.conversion ?? DEFAULT_OBJECTIVE_WEIGHTS.conversion),
    ltv: String(config?.weights.ltv ?? DEFAULT_OBJECTIVE_WEIGHTS.ltv),
    revenue: String(config?.weights.revenue ?? DEFAULT_OBJECTIVE_WEIGHTS.revenue),
  };
}

function mergeArmStats(experiment: ExperimentSummary | null, snapshot: MultiObjectiveSnapshot | null) {
  if (!experiment) return [];

  const totalWeight = experiment.arms.reduce((sum, arm) => sum + Math.max(arm.traffic_weight, 0), 0);
  const statsByArmId = new Map<string, BanditArmStatistics>();
  for (const stats of snapshot?.banditSnapshot?.statistics?.arms ?? []) {
    statsByArmId.set(stats.arm_id, stats);
  }

  return experiment.arms.map((arm) => {
    const stats = statsByArmId.get(arm.id);
    const conversionRate =
      stats?.conversion_rate ??
      (stats && stats.samples > 0
        ? stats.conversions / stats.samples
        : arm.samples > 0
          ? arm.conversions / arm.samples
          : 0);
    return {
      ...arm,
      samples: stats?.samples ?? arm.samples,
      conversions: stats?.conversions ?? arm.conversions,
      revenue: stats?.revenue ?? arm.revenue,
      conversionRate,
      weightShare: totalWeight > 0 ? arm.traffic_weight / totalWeight : 0,
      winProbability: snapshot?.banditSnapshot?.statistics?.win_probabilities?.[arm.id] ?? null,
      objectiveScores: snapshot?.objectiveScores?.[arm.id] ?? {},
    };
  });
}

export function MultiObjectivePageClient({
  initialExperiments,
  initialSelectedExperimentId = null,
  initialSnapshot = null,
  loadFailed: initialLoadFailed = false,
}: {
  initialExperiments?: ExperimentSummary[];
  initialSelectedExperimentId?: string | null;
  initialSnapshot?: MultiObjectiveSnapshot | null;
  loadFailed?: boolean;
}) {
  const hasInitialPayload = initialExperiments !== undefined;
  const t = useTranslations("multiObjective");
  const [experiments, setExperiments] = useState<ExperimentSummary[]>(initialExperiments ?? []);
  const [selectedId, setSelectedId] = useState(initialSelectedExperimentId ?? "");
  const [snapshot, setSnapshot] = useState<MultiObjectiveSnapshot | null>(initialSnapshot ?? null);
  const [loadFailed, setLoadFailed] = useState(initialLoadFailed);
  const [isBootstrapping, setIsBootstrapping] = useState(!hasInitialPayload);
  const [isPending, startTransition] = useTransition();
  const [objectiveType, setObjectiveType] = useState<ObjectiveType>(
    initialSnapshot?.currentConfig?.objectiveType ?? DEFAULT_OBJECTIVE_TYPE,
  );
  const [weightInputs, setWeightInputs] = useState(buildWeightInputs(initialSnapshot?.currentConfig));
  const [isSavingConfig, setIsSavingConfig] = useState(false);
  const [configResult, setConfigResult] = useState<ObjectiveConfigResult | null>(null);

  useEffect(() => {
    if (!isBootstrapping) return;

    startTransition(async () => {
      try {
        const data = await fetchMultiObjectiveJson<MultiObjectiveDashboardData>("/api/admin/multi-objective/dashboard");
        setExperiments(data.experiments);
        setSelectedId(data.selectedExperimentId ?? "");
        setSnapshot(data.snapshot);
        setLoadFailed(data.loadFailed);
      } catch {
        setLoadFailed(true);
      } finally {
        setIsBootstrapping(false);
      }
    });
  }, [isBootstrapping]);

  useEffect(() => {
    setObjectiveType(snapshot?.currentConfig?.objectiveType ?? DEFAULT_OBJECTIVE_TYPE);
    setWeightInputs(buildWeightInputs(snapshot?.currentConfig));
  }, [snapshot]);

  const selectedExperiment = useMemo(
    () => experiments.find((experiment) => experiment.id === selectedId) ?? snapshot?.experiment ?? null,
    [experiments, selectedId, snapshot],
  );
  const armRows = useMemo(() => mergeArmStats(selectedExperiment, snapshot), [selectedExperiment, snapshot]);
  const totalSamples = armRows.reduce((sum, arm) => sum + arm.samples, 0);
  const totalRevenue = armRows.reduce((sum, arm) => sum + arm.revenue, 0);
  const objectiveRouteReady = snapshot?.probes.objectiveScores.state === "available";

  const refreshSnapshot = (experimentId: string) => {
    setSelectedId(experimentId);
    setConfigResult(null);
    startTransition(async () => {
      try {
        const data = await fetchMultiObjectiveJson<MultiObjectiveSnapshot>(
          `/api/admin/multi-objective/snapshot?experimentId=${encodeURIComponent(experimentId)}`,
        );
        setSnapshot(data);
      } catch {
        toast.error(t("feedback.loadFailed"));
      }
    });
  };

  const hybridWeightsValid = [weightInputs.conversion, weightInputs.ltv, weightInputs.revenue].every((value) => {
    const parsed = Number(value);
    return Number.isFinite(parsed) && parsed >= 0;
  });

  const saveObjectiveConfig = async () => {
    if (!selectedId || (objectiveType === "hybrid" && !hybridWeightsValid)) return;
    setIsSavingConfig(true);

    try {
      const res = await fetch("/api/admin/multi-objective/config", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          experimentId: selectedId,
          objectiveType,
          objectiveWeights:
            objectiveType === "hybrid"
              ? {
                  conversion: Number(weightInputs.conversion),
                  ltv: Number(weightInputs.ltv),
                  revenue: Number(weightInputs.revenue),
                }
              : undefined,
        }),
      });
      const body = await res.json().catch(() => ({}));
      const result = {
        ok: res.ok,
        status: res.status,
        message:
          (body as { message?: string; error?: string }).message ??
          (body as { error?: string }).error ??
          `HTTP ${res.status}`,
        objectiveType: ((body as { objective_type?: ObjectiveType }).objective_type ??
          (body as { data?: { objective_type?: ObjectiveType } }).data?.objective_type ??
          objectiveType) as ObjectiveType,
        weights:
          (body as { weights?: Record<string, number> }).weights ??
          (body as { data?: { weights?: Record<string, number> } }).data?.weights,
      } satisfies ObjectiveConfigResult;

      setConfigResult(result);
      if (!result.ok) {
        toast.error(result.message);
        return;
      }

      if (result.weights) {
        setWeightInputs({
          conversion: String(result.weights.conversion ?? 0),
          ltv: String(result.weights.ltv ?? 0),
          revenue: String(result.weights.revenue ?? 0),
        });
      }
      setObjectiveType(result.objectiveType ?? objectiveType);
      toast.success(t("feedback.configSaved"));
      refreshSnapshot(selectedId);
    } catch {
      const result = {
        ok: false,
        status: 500,
        message: t("feedback.saveFailed"),
        objectiveType,
      } satisfies ObjectiveConfigResult;
      setConfigResult(result);
      toast.error(result.message);
    } finally {
      setIsSavingConfig(false);
    }
  };

  const probeRows = snapshot
    ? [
        { key: "objectiveScores", probe: snapshot.probes.objectiveScores },
        { key: "objectiveConfig", probe: snapshot.probes.objectiveConfig },
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
                <FlaskConical className="size-4" />
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
                  <Badge variant="outline">{formatAlgorithm(selectedExperiment.algorithm_type)}</Badge>
                </div>
              ) : null}
            </CardContent>
          </Card>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {[
              {
                label: t("summary.runtime"),
                value: snapshot?.serviceHealth
                  ? `${snapshot.serviceHealth.service}: ${snapshot.serviceHealth.status}`
                  : t("summary.unavailable"),
              },
              { label: t("summary.samples"), value: totalSamples.toLocaleString("en-US") },
              { label: t("summary.revenue"), value: formatCurrency(totalRevenue) },
              {
                label: t("summary.objectiveRoute"),
                value: objectiveRouteReady ? t("summary.available") : t("summary.unavailable"),
              },
            ].map((item) => (
              <Card key={item.label}>
                <CardContent className="pt-6">
                  <p className="font-semibold text-muted-foreground text-xs uppercase tracking-widest">{item.label}</p>
                  <p className="mt-2 break-all font-bold text-lg">{item.value}</p>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.1fr_1fr]">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">{t("probes.title")}</CardTitle>
                <CardDescription>{t("probes.description")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {probeRows.map((item) => (
                  <div key={item.key} className="rounded-md border p-3">
                    <div className="flex items-center justify-between gap-3">
                      <p className="font-medium text-sm">{t(`probes.${item.key}.label`)}</p>
                      <Badge className={probeClass(item.probe.state)}>{t(`probes.states.${item.probe.state}`)}</Badge>
                    </div>
                    <p className="mt-1 text-muted-foreground text-xs">
                      {item.probe.status ? `HTTP ${item.probe.status} · ` : ""}
                      {item.probe.message}
                    </p>
                  </div>
                ))}
              </CardContent>
            </Card>

            <div className="flex flex-col gap-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("config.title")}</CardTitle>
                  <CardDescription>{t("config.description")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  {snapshot?.currentConfig ? (
                    <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                      <p className="font-medium text-foreground text-xs">{t("config.loadedTitle")}</p>
                      <p className="mt-1">
                        {t("config.loadedBody", {
                          objective: t(`catalog.objectives.${snapshot.currentConfig.objectiveType}.label`),
                        })}
                      </p>
                    </div>
                  ) : null}

                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("config.objectiveType")}</p>
                    <Select value={objectiveType} onValueChange={(value) => setObjectiveType(value as ObjectiveType)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {SUPPORTED_OBJECTIVES.map((objective) => (
                          <SelectItem key={objective.key} value={objective.key}>
                            {t(`catalog.objectives.${objective.key}.label`)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="grid gap-3 sm:grid-cols-3">
                    {(["conversion", "ltv", "revenue"] as const).map((key) => (
                      <div key={key} className="space-y-1">
                        <p className="font-medium text-xs">{t(`config.weights.${key}`)}</p>
                        <Input
                          value={weightInputs[key]}
                          onChange={(event) =>
                            setWeightInputs((current) => ({ ...current, [key]: event.target.value }))
                          }
                          disabled={objectiveType !== "hybrid"}
                          inputMode="decimal"
                        />
                      </div>
                    ))}
                  </div>

                  <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                    {t(objectiveType === "hybrid" ? "config.hybridHint" : "config.singleHint")}
                  </div>

                  <Button
                    onClick={saveObjectiveConfig}
                    disabled={!selectedId || isSavingConfig || (objectiveType === "hybrid" && !hybridWeightsValid)}
                  >
                    {isSavingConfig ? t("actions.savingConfig") : t("actions.saveConfig")}
                  </Button>

                  {configResult ? (
                    <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                      {configResult.status ? `HTTP ${configResult.status} · ` : ""}
                      {configResult.message}
                    </div>
                  ) : null}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("catalog.title")}</CardTitle>
                  <CardDescription>{t("catalog.description")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  {SUPPORTED_OBJECTIVES.map((objective) => (
                    <div key={objective.key} className="rounded-md border p-3">
                      <div className="flex items-center justify-between gap-3">
                        <p className="font-medium text-sm">{t(`catalog.objectives.${objective.key}.label`)}</p>
                        {objective.weightPercent === null ? (
                          <Badge variant="outline">{t("catalog.derived")}</Badge>
                        ) : (
                          <Badge variant="outline">{objective.weightPercent}%</Badge>
                        )}
                      </div>
                      <p className="mt-1 text-muted-foreground text-xs">
                        {t(`catalog.objectives.${objective.key}.body`)}
                      </p>
                      {objective.weightPercent !== null ? (
                        <div className="mt-2 space-y-1">
                          <Progress value={objective.weightPercent} className="h-2" />
                          <p className="text-muted-foreground text-xs">
                            {t("catalog.defaultHybridWeight")} {objective.defaultWeight?.toFixed(1)}
                          </p>
                        </div>
                      ) : null}
                    </div>
                  ))}
                </CardContent>
              </Card>
            </div>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("context.title")}</CardTitle>
              <CardDescription>{t("context.description")}</CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("table.arm")}</TableHead>
                    <TableHead>{t("table.weight")}</TableHead>
                    <TableHead>{t("table.samples")}</TableHead>
                    <TableHead>{t("table.conversions")}</TableHead>
                    <TableHead>{t("table.conversionRate")}</TableHead>
                    <TableHead>{t("table.winProbability")}</TableHead>
                    <TableHead>{t("table.revenue")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {armRows.map((arm) => (
                    <TableRow key={arm.id}>
                      <TableCell>
                        <div>
                          <p className="font-medium">{arm.name}</p>
                          <p className="text-muted-foreground text-xs">
                            {arm.is_control ? t("table.control") : t("table.variant")}
                          </p>
                        </div>
                      </TableCell>
                      <TableCell className="min-w-36">
                        <div className="space-y-1">
                          <Progress value={arm.weightShare * 100} className="h-2" />
                          <p className="text-muted-foreground text-xs">
                            {arm.traffic_weight.toFixed(2)} · {formatPercent(arm.weightShare, 0)}
                          </p>
                        </div>
                      </TableCell>
                      <TableCell>{arm.samples.toLocaleString("en-US")}</TableCell>
                      <TableCell>{arm.conversions.toLocaleString("en-US")}</TableCell>
                      <TableCell>{formatRatePercent(arm.conversionRate)}</TableCell>
                      <TableCell>{formatPercent(arm.winProbability)}</TableCell>
                      <TableCell>{formatCurrency(arm.revenue)}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("scores.title")}</CardTitle>
              <CardDescription>{t("scores.description")}</CardDescription>
            </CardHeader>
            <CardContent>
              {snapshot?.objectiveScores ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t("scores.table.arm")}</TableHead>
                      <TableHead>{t("scores.table.conversion")}</TableHead>
                      <TableHead>{t("scores.table.ltv")}</TableHead>
                      <TableHead>{t("scores.table.revenue")}</TableHead>
                      <TableHead>{t("scores.table.hybrid")}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {armRows.map((arm) => (
                      <TableRow key={arm.id}>
                        <TableCell>{arm.name}</TableCell>
                        <TableCell className="font-mono">
                          {formatScore(arm.objectiveScores.conversion?.score)}
                        </TableCell>
                        <TableCell className="font-mono">{formatScore(arm.objectiveScores.ltv?.score)}</TableCell>
                        <TableCell className="font-mono">{formatScore(arm.objectiveScores.revenue?.score)}</TableCell>
                        <TableCell className="font-mono">{formatScore(arm.objectiveScores.hybrid?.score)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <div className="rounded-md border border-dashed p-4 text-muted-foreground text-sm">
                  {t("scores.unavailable")}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("notes.title")}</CardTitle>
              <CardDescription>{t("notes.description")}</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-3 text-sm md:grid-cols-3">
              <div className="rounded-md border p-3">
                <p className="font-medium">{t("notes.strategyTitle")}</p>
                <p className="mt-1 text-muted-foreground text-xs">{t("notes.strategyBody")}</p>
              </div>
              <div className="rounded-md border p-3">
                <p className="font-medium">{t("notes.runtimeTitle")}</p>
                <p className="mt-1 text-muted-foreground text-xs">{t("notes.runtimeBody")}</p>
              </div>
              <div className="rounded-md border p-3">
                <p className="font-medium">{t("notes.configTitle")}</p>
                <p className="mt-1 text-muted-foreground text-xs">{t("notes.configBody")}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("readiness.title")}</CardTitle>
              <CardDescription>{t("readiness.description")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-center justify-between rounded-md border p-3">
                <span>{t("readiness.objectiveScores")}</span>
                <span className="font-mono">
                  {snapshot?.probes.objectiveScores.state === "available" ? "1/1" : "0/1"}
                </span>
              </div>
              <div className="flex items-center justify-between rounded-md border p-3">
                <span>{t("readiness.topArm")}</span>
                <span className="font-mono">
                  {armRows.length > 0
                    ? (armRows
                        .slice()
                        .sort((left, right) => (right.winProbability ?? -1) - (left.winProbability ?? -1))[0]?.name ??
                      "—")
                    : "—"}
                </span>
              </div>
              <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                <Target className="mb-2 size-4" />
                {t("readiness.body")}
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
