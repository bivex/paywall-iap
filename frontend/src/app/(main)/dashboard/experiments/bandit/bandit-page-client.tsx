"use client";

import { useEffect, useMemo, useState, useTransition } from "react";

import Link from "next/link";

import { Brain, FlaskConical, RefreshCw } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { BanditArmStatistics, BanditSnapshot } from "@/lib/bandit";
import type { ExperimentAlgorithm, ExperimentStatus, ExperimentSummary } from "@/lib/experiments";

function formatPercent(value: number | null | undefined, digits = 1) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return `${(value * 100).toFixed(digits)}%`;
}

function formatRatePercent(value: number | null | undefined, digits = 2) {
  if (value === null || value === undefined || !Number.isFinite(value)) return "—";
  return `${(value * 100).toFixed(digits)}%`;
}

function formatRevenue(value: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
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

function mergeArmStats(experiment: ExperimentSummary | null, snapshot: BanditSnapshot | null) {
  if (!experiment) return [];

  const totalWeight = experiment.arms.reduce((sum, arm) => sum + Math.max(arm.traffic_weight, 0), 0);
  const statsByArmId = new Map<string, BanditArmStatistics>();
  for (const stats of snapshot?.statistics?.arms ?? []) {
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
      alpha: stats?.alpha ?? 1,
      beta: stats?.beta ?? 1,
      samples: stats?.samples ?? arm.samples,
      conversions: stats?.conversions ?? arm.conversions,
      revenue: stats?.revenue ?? arm.revenue,
      avgReward: stats?.avg_reward ?? arm.avg_reward,
      conversionRate,
      weightShare: totalWeight > 0 ? arm.traffic_weight / totalWeight : 0,
      winProbability: snapshot?.statistics?.win_probabilities?.[arm.id] ?? null,
    };
  });
}

async function fetchBanditJson<T>(url: string): Promise<T> {
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

export function BanditPageClient({
  initialExperiments = [],
  initialSelectedExperimentId = null,
  initialSnapshot = null,
  loadFailed: initialLoadFailed = false,
}: {
  initialExperiments?: ExperimentSummary[];
  initialSelectedExperimentId?: string | null;
  initialSnapshot?: BanditSnapshot | null;
  loadFailed?: boolean;
}) {
  const t = useTranslations("bandit");
  const [experiments, setExperiments] = useState(initialExperiments);
  const [selectedId, setSelectedId] = useState(initialSelectedExperimentId ?? "");
  const [snapshot, setSnapshot] = useState(initialSnapshot);
  const [loadFailed, setLoadFailed] = useState(initialLoadFailed);
  const [isBootstrapping, setIsBootstrapping] = useState(
    initialExperiments.length === 0 &&
      initialSelectedExperimentId === null &&
      initialSnapshot === null &&
      !initialLoadFailed,
  );
  const [isPending, startTransition] = useTransition();

  useEffect(() => {
    if (!isBootstrapping) return;

    startTransition(async () => {
      try {
        const data = await fetchBanditJson<{
          experiments: ExperimentSummary[];
          selectedExperimentId: string | null;
          snapshot: BanditSnapshot | null;
          loadFailed: boolean;
        }>("/api/admin/bandit/dashboard");
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

  const selectedExperiment = useMemo(
    () => experiments.find((experiment) => experiment.id === selectedId) ?? null,
    [experiments, selectedId],
  );
  const armRows = useMemo(() => mergeArmStats(selectedExperiment, snapshot), [selectedExperiment, snapshot]);
  const topArm = useMemo(
    () => [...armRows].sort((left, right) => (right.winProbability ?? -1) - (left.winProbability ?? -1))[0] ?? null,
    [armRows],
  );

  const totalSamples = armRows.reduce((sum, arm) => sum + arm.samples, 0);
  const totalConversions = armRows.reduce((sum, arm) => sum + arm.conversions, 0);
  const totalRevenue = armRows.reduce((sum, arm) => sum + arm.revenue, 0);

  const loadSnapshot = (experimentId: string) => {
    setSelectedId(experimentId);
    setSnapshot(null);
    startTransition(async () => {
      try {
        const nextSnapshot = await fetchBanditJson<BanditSnapshot>(
          `/api/admin/bandit/snapshot?experimentId=${encodeURIComponent(experimentId)}`,
        );
        setSnapshot(nextSnapshot);
      } catch {
        toast.error(t("feedback.loadFailed"));
      }
    });
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-semibold text-2xl tracking-tight">{t("title")}</h1>
          <p className="mt-0.5 text-muted-foreground text-sm">{t("subtitle")}</p>
        </div>

        {selectedExperiment ? (
          <div className="flex items-center gap-2">
            <Badge className={statusClass(selectedExperiment.status)}>{t(`status.${selectedExperiment.status}`)}</Badge>
            <Badge variant="outline">{formatAlgorithm(selectedExperiment.algorithm_type)}</Badge>
          </div>
        ) : null}
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
              <Select value={selectedId} onValueChange={loadSnapshot}>
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

              <Button
                variant="outline"
                size="sm"
                onClick={() => selectedId && loadSnapshot(selectedId)}
                disabled={!selectedId || isPending}
              >
                <RefreshCw className={`size-4 ${isPending ? "animate-spin" : ""}`} />
                {t("actions.refresh")}
              </Button>
            </CardContent>
          </Card>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-5">
            {[
              { label: t("summary.samples"), value: totalSamples.toLocaleString("en-US") },
              { label: t("summary.conversions"), value: totalConversions.toLocaleString("en-US") },
              { label: t("summary.revenue"), value: formatRevenue(totalRevenue) },
              { label: t("summary.leadingArm"), value: topArm?.name ?? "—" },
              {
                label: t("summary.balanceIndex"),
                value: snapshot?.metrics ? formatPercent(snapshot.metrics.balance_index, 0) : t("summary.unavailable"),
              },
            ].map((item) => (
              <Card key={item.label}>
                <CardContent className="pt-6">
                  <p className="font-semibold text-muted-foreground text-xs uppercase tracking-widest">{item.label}</p>
                  <p className="mt-2 font-bold text-xl tabular-nums">{item.value}</p>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.5fr_1fr]">
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
                        <TableCell>{formatRevenue(arm.revenue)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>

            <div className="flex flex-col gap-4">
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("metrics.title")}</CardTitle>
                  <CardDescription>{t("metrics.description")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <div className="flex items-center justify-between rounded-md border p-3">
                    <span>{t("metrics.balanceIndex")}</span>
                    <span className="font-mono">
                      {snapshot?.metrics ? formatPercent(snapshot.metrics.balance_index) : "—"}
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-md border p-3">
                    <span>{t("metrics.pendingRewards")}</span>
                    <span className="font-mono">{snapshot?.metrics?.pending_rewards ?? 0}</span>
                  </div>
                  <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                    {t("metrics.note")}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("config.title")}</CardTitle>
                  <CardDescription>{t("config.description")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-2 text-sm">
                  <div className="flex items-center justify-between rounded-md border p-3">
                    <span>{t("config.algorithm")}</span>
                    <span className="font-mono">{formatAlgorithm(selectedExperiment?.algorithm_type ?? null)}</span>
                  </div>
                  <div className="flex items-center justify-between rounded-md border p-3">
                    <span>{t("config.minSampleSize")}</span>
                    <span className="font-mono">{selectedExperiment?.min_sample_size ?? "—"}</span>
                  </div>
                  <div className="flex items-center justify-between rounded-md border p-3">
                    <span>{t("config.confidence")}</span>
                    <span className="font-mono">
                      {selectedExperiment ? `${selectedExperiment.confidence_threshold_percent.toFixed(0)}%` : "—"}
                    </span>
                  </div>
                  <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                    <Brain className="mb-2 size-4" />
                    {t("config.readOnlyNotice")}
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
