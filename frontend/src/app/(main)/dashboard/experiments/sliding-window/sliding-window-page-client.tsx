"use client";

import { useEffect, useMemo, useState, useTransition } from "react";

import Link from "next/link";

import { BarChart2, Brain, Download, FlaskConical, RefreshCw } from "lucide-react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import type { BanditArmStatistics } from "@/lib/bandit";
import type { ExperimentAlgorithm, ExperimentStatus, ExperimentSummary } from "@/lib/experiments";
import type {
  SlidingWindowDashboardData,
  SlidingWindowEndpointProbe,
  SlidingWindowEventsExport,
  SlidingWindowRewardEvent,
  SlidingWindowSnapshot,
  TrimWindowResult,
} from "@/lib/sliding-window";

const WINDOW_EVENT_LIMITS = [25, 50, 100, 250] as const;

async function fetchSlidingWindowJson<T>(url: string): Promise<T> {
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

function probeClass(state: SlidingWindowEndpointProbe["state"]) {
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

function formatRevenue(value: number) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

function formatDateTime(value: string | null | undefined) {
  if (!value) return "—";

  try {
    return new Date(value).toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    });
  } catch {
    return value;
  }
}

function formatEventReward(value: number, currency: string) {
  if (currency) {
    try {
      return new Intl.NumberFormat("en-US", {
        style: "currency",
        currency,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(value);
    } catch {
      return `${value.toFixed(2)} ${currency}`.trim();
    }
  }

  return value.toFixed(2);
}

function formatDurationNs(value: number | null | undefined) {
  if (value === null || value === undefined || !Number.isFinite(value) || value <= 0) return "—";

  const milliseconds = value / 1_000_000;
  if (milliseconds < 1000) return `${milliseconds.toFixed(0)} ms`;

  const seconds = milliseconds / 1000;
  if (seconds < 60) return `${seconds.toFixed(1)} s`;

  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${(seconds % 60).toFixed(0)}s`;
}

function formatMetadata(metadata: Record<string, unknown> | null | undefined) {
  if (!metadata || Object.keys(metadata).length === 0) return "—";

  try {
    return JSON.stringify(metadata);
  } catch {
    return "[unserializable metadata]";
  }
}

interface SlidingWindowEventsResult {
  ok: boolean;
  status: number;
  message: string;
  data?: SlidingWindowEventsExport;
}

function mergeArmStats(experiment: ExperimentSummary | null, snapshot: SlidingWindowSnapshot | null) {
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
    };
  });
}

export function SlidingWindowPageClient({
  initialExperiments,
  initialSelectedExperimentId = null,
  initialSnapshot = null,
  loadFailed: initialLoadFailed = false,
}: {
  initialExperiments?: ExperimentSummary[];
  initialSelectedExperimentId?: string | null;
  initialSnapshot?: SlidingWindowSnapshot | null;
  loadFailed?: boolean;
}) {
  const hasInitialPayload = initialExperiments !== undefined;
  const t = useTranslations("slidingWindow");
  const [experiments, setExperiments] = useState<ExperimentSummary[]>(initialExperiments ?? []);
  const [selectedId, setSelectedId] = useState(initialSelectedExperimentId ?? "");
  const [snapshot, setSnapshot] = useState<SlidingWindowSnapshot | null>(initialSnapshot ?? null);
  const [loadFailed, setLoadFailed] = useState(initialLoadFailed);
  const [isBootstrapping, setIsBootstrapping] = useState(!hasInitialPayload);
  const [isPending, startTransition] = useTransition();
  const [isTrimming, setIsTrimming] = useState(false);
  const [trimResult, setTrimResult] = useState<TrimWindowResult | null>(null);
  const [eventsLimit, setEventsLimit] = useState<string>(String(WINDOW_EVENT_LIMITS[2]));
  const [isLoadingEvents, setIsLoadingEvents] = useState(false);
  const [eventsResult, setEventsResult] = useState<SlidingWindowEventsResult | null>(null);

  useEffect(() => {
    if (!isBootstrapping) return;

    startTransition(async () => {
      try {
        const data = await fetchSlidingWindowJson<SlidingWindowDashboardData>("/api/admin/sliding-window/dashboard");
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
    () => experiments.find((experiment) => experiment.id === selectedId) ?? snapshot?.experiment ?? null,
    [experiments, selectedId, snapshot],
  );
  const armRows = useMemo(() => mergeArmStats(selectedExperiment, snapshot), [selectedExperiment, snapshot]);
  const totalSamples = armRows.reduce((sum, arm) => sum + arm.samples, 0);
  const totalRevenue = armRows.reduce((sum, arm) => sum + arm.revenue, 0);
  const availableReadRoutes = snapshot
    ? [snapshot.probes.windowInfo, snapshot.probes.windowEvents].filter((probe) => probe.state === "available").length
    : 0;
  const selectedArmNames = useMemo(
    () => new Map((selectedExperiment?.arms ?? []).map((arm) => [arm.id, arm.name])),
    [selectedExperiment],
  );
  const groupedEvents = useMemo(
    () =>
      Object.entries(eventsResult?.data?.events ?? {}).map(([armId, events]) => ({
        armId,
        armName: selectedArmNames.get(armId) ?? armId,
        events: events as SlidingWindowRewardEvent[],
      })),
    [eventsResult, selectedArmNames],
  );
  const totalExportedEvents = groupedEvents.reduce((sum, group) => sum + group.events.length, 0);

  const refreshSnapshot = (experimentId: string) => {
    setSelectedId(experimentId);
    setTrimResult(null);
    setEventsResult(null);
    startTransition(async () => {
      try {
        const data = await fetchSlidingWindowJson<SlidingWindowSnapshot>(
          `/api/admin/sliding-window/snapshot?experimentId=${encodeURIComponent(experimentId)}`,
        );
        setSnapshot(data);
      } catch {
        toast.error(t("feedback.loadFailed"));
      }
    });
  };

  const trimWindow = async () => {
    if (!selectedId) return;
    setIsTrimming(true);

    try {
      const res = await fetch("/api/admin/sliding-window/trim", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ experimentId: selectedId }),
      });
      const body = await res.json().catch(() => ({}));
      const result = {
        ok: res.ok,
        status: res.status,
        message:
          (body as { message?: string; error?: string }).message ??
          (body as { error?: string }).error ??
          `HTTP ${res.status}`,
        experimentId: selectedId,
      } satisfies TrimWindowResult;

      setTrimResult(result);
      if (!result.ok) {
        toast.error(result.message);
        return;
      }

      toast.success(t("feedback.trimmed"));
      refreshSnapshot(selectedId);
    } catch {
      const result = {
        ok: false,
        status: 500,
        message: t("feedback.trimFailed"),
        experimentId: selectedId,
      } satisfies TrimWindowResult;
      setTrimResult(result);
      toast.error(result.message);
    } finally {
      setIsTrimming(false);
    }
  };

  const loadWindowEvents = async () => {
    if (!selectedId) return;

    setIsLoadingEvents(true);
    try {
      const params = new URLSearchParams({
        experimentId: selectedId,
        limit: eventsLimit,
      });
      const res = await fetch(`/api/admin/sliding-window/events?${params.toString()}`, {
        cache: "no-store",
      });
      const body = await res.json().catch(() => ({}));
      const result = {
        ok: res.ok,
        status: res.status,
        message:
          (body as { message?: string; error?: string }).message ??
          (body as { error?: string }).error ??
          `HTTP ${res.status}`,
        data: res.ok
          ? (((body as { data?: SlidingWindowEventsExport }).data ?? body) as SlidingWindowEventsExport)
          : undefined,
      } satisfies SlidingWindowEventsResult;

      setEventsResult(result);
      if (!result.ok) {
        toast.error(result.message);
        return;
      }

      toast.success(t("feedback.eventsLoaded"));
    } catch {
      const result = {
        ok: false,
        status: 500,
        message: t("feedback.eventsFailed"),
      } satisfies SlidingWindowEventsResult;
      setEventsResult(result);
      toast.error(result.message);
    } finally {
      setIsLoadingEvents(false);
    }
  };

  const downloadWindowEvents = () => {
    if (!eventsResult?.ok || !eventsResult.data || typeof window === "undefined") return;

    const blob = new Blob([JSON.stringify(eventsResult.data, null, 2)], { type: "application/json" });
    const url = window.URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `sliding-window-events-${eventsResult.data.experiment_id}-${eventsResult.data.limit}.json`;
    document.body.append(anchor);
    anchor.click();
    anchor.remove();
    window.URL.revokeObjectURL(url);
  };

  const probeRows = snapshot
    ? [
        { key: "windowInfo", probe: snapshot.probes.windowInfo },
        { key: "windowEvents", probe: snapshot.probes.windowEvents },
        { key: "trimWindow", probe: snapshot.probes.trimWindow },
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
              { label: t("summary.revenue"), value: formatRevenue(totalRevenue) },
              { label: t("summary.readRoutes"), value: `${availableReadRoutes}/2` },
            ].map((item) => (
              <Card key={item.label}>
                <CardContent className="pt-6">
                  <p className="font-semibold text-muted-foreground text-xs uppercase tracking-widest">{item.label}</p>
                  <p className="mt-2 break-all font-bold text-lg">{item.value}</p>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1.15fr_1fr]">
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
                  <CardTitle className="text-sm">{t("trim.title")}</CardTitle>
                  <CardDescription>{t("trim.description")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <div className="rounded-md border p-3">
                    <p className="font-medium">{selectedExperiment?.name ?? "—"}</p>
                    <p className="mt-1 text-muted-foreground text-xs">{t("trim.body")}</p>
                  </div>
                  <Button onClick={trimWindow} disabled={!selectedId || isPending || isTrimming}>
                    {isTrimming ? t("actions.trimming") : t("actions.trimWindow")}
                  </Button>
                  {trimResult ? (
                    <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                      {trimResult.status ? `HTTP ${trimResult.status} · ` : ""}
                      {trimResult.message}
                    </div>
                  ) : null}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("notes.title")}</CardTitle>
                  <CardDescription>{t("notes.description")}</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3 text-sm">
                  <div className="rounded-md border p-3">
                    <p className="font-medium">{t("notes.strategyTitle")}</p>
                    <p className="mt-1 text-muted-foreground text-xs">{t("notes.strategyBody")}</p>
                  </div>
                  <div className="rounded-md border p-3">
                    <p className="font-medium">{t("notes.readRoutesTitle")}</p>
                    <p className="mt-1 text-muted-foreground text-xs">{t("notes.readRoutesBody")}</p>
                  </div>
                  <div className="rounded-md border p-3">
                    <p className="font-medium">{t("notes.trimTitle")}</p>
                    <p className="mt-1 text-muted-foreground text-xs">{t("notes.trimBody")}</p>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("events.title")}</CardTitle>
              <CardDescription>{t("events.description")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-wrap items-end gap-3">
                <div className="space-y-1.5">
                  <p className="font-medium text-sm">{t("events.limit")}</p>
                  <Select value={eventsLimit} onValueChange={setEventsLimit}>
                    <SelectTrigger className="w-[160px]">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {WINDOW_EVENT_LIMITS.map((limit) => (
                        <SelectItem key={limit} value={String(limit)}>
                          {t("events.limitOption", { limit })}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button onClick={loadWindowEvents} disabled={!selectedId || isLoadingEvents}>
                    {isLoadingEvents ? t("events.loading") : t("events.load")}
                  </Button>
                  <Button
                    variant="outline"
                    onClick={downloadWindowEvents}
                    disabled={!eventsResult?.ok || !eventsResult.data}
                  >
                    <Download className="size-4" />
                    {t("events.download")}
                  </Button>
                </div>
              </div>

              <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                {t("events.body")}
              </div>

              {eventsResult ? (
                <div className="rounded-md border p-3 text-sm">
                  <p className="font-medium">{t("events.httpStatus", { status: eventsResult.status })}</p>
                  <p className={`mt-1 text-xs ${eventsResult.ok ? "text-muted-foreground" : "text-destructive"}`}>
                    {eventsResult.message}
                  </p>

                  {eventsResult.ok && eventsResult.data ? (
                    <>
                      <div className="mt-3 grid gap-3 md:grid-cols-3">
                        <div className="rounded-md border p-3">
                          <p className="text-muted-foreground text-xs">{t("events.summary.experimentId")}</p>
                          <p className="mt-1 break-all font-medium">{eventsResult.data.experiment_id}</p>
                        </div>
                        <div className="rounded-md border p-3">
                          <p className="text-muted-foreground text-xs">{t("events.summary.limit")}</p>
                          <p className="mt-1 font-medium">{eventsResult.data.limit}</p>
                        </div>
                        <div className="rounded-md border p-3">
                          <p className="text-muted-foreground text-xs">{t("events.summary.total")}</p>
                          <p className="mt-1 font-medium">{totalExportedEvents}</p>
                        </div>
                      </div>

                      {groupedEvents.length === 0 ? (
                        <p className="mt-3 text-muted-foreground text-xs">{t("events.empty")}</p>
                      ) : (
                        <div className="mt-4 space-y-4">
                          {groupedEvents.map((group) => (
                            <div key={group.armId} className="rounded-md border p-3">
                              <div className="flex flex-wrap items-center justify-between gap-2">
                                <div>
                                  <p className="font-medium">{group.armName}</p>
                                  <p className="mt-1 break-all text-muted-foreground text-xs">{group.armId}</p>
                                </div>
                                <Badge variant="outline">
                                  {t("events.summary.groupCount", { count: group.events.length })}
                                </Badge>
                              </div>

                              <div className="mt-3 overflow-x-auto">
                                <Table>
                                  <TableHeader>
                                    <TableRow>
                                      <TableHead>{t("events.table.user")}</TableHead>
                                      <TableHead>{t("events.table.timestamp")}</TableHead>
                                      <TableHead>{t("events.table.reward")}</TableHead>
                                      <TableHead>{t("events.table.delay")}</TableHead>
                                      <TableHead>{t("events.table.metadata")}</TableHead>
                                    </TableRow>
                                  </TableHeader>
                                  <TableBody>
                                    {group.events.map((event, index) => {
                                      const metadata = formatMetadata(event.Metadata);

                                      return (
                                        <TableRow key={`${group.armId}-${event.UserID}-${event.Timestamp}-${index}`}>
                                          <TableCell className="max-w-44 break-all text-xs">
                                            {event.UserID || "—"}
                                          </TableCell>
                                          <TableCell className="text-xs">{formatDateTime(event.Timestamp)}</TableCell>
                                          <TableCell className="text-xs">
                                            {formatEventReward(event.RewardValue, event.Currency)}
                                          </TableCell>
                                          <TableCell className="text-xs">
                                            {formatDurationNs(event.ConversionDelay)}
                                          </TableCell>
                                          <TableCell className="max-w-72 truncate text-xs" title={metadata}>
                                            {metadata}
                                          </TableCell>
                                        </TableRow>
                                      );
                                    })}
                                  </TableBody>
                                </Table>
                              </div>
                            </div>
                          ))}
                        </div>
                      )}
                    </>
                  ) : null}
                </div>
              ) : null}
            </CardContent>
          </Card>

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

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("readiness.title")}</CardTitle>
              <CardDescription>{t("readiness.description")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex items-center justify-between rounded-md border p-3">
                <span>{t("readiness.readRoutes")}</span>
                <span className="font-mono">{availableReadRoutes}/2</span>
              </div>
              <div className="flex items-center justify-between rounded-md border p-3">
                <span>{t("readiness.topArm")}</span>
                <span className="font-mono">
                  {armRows[0]
                    ? `${armRows.slice().sort((left, right) => (right.winProbability ?? -1) - (left.winProbability ?? -1))[0]?.name ?? "—"}`
                    : "—"}
                </span>
              </div>
              <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                <BarChart2 className="mb-2 size-4" />
                {t("readiness.body")}
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
