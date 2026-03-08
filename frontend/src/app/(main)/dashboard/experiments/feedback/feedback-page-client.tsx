"use client";

import { useEffect, useMemo, useState, useTransition } from "react";

import Link from "next/link";

import { Activity, Brain, FlaskConical, RefreshCw, Send } from "lucide-react";
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
import type {
  DelayedConversionPayload,
  DelayedConversionResult,
  DelayedEndpointProbe,
  DelayedFeedbackDashboardData,
  DelayedPendingReward,
  DelayedPendingRewardsByUser,
  DelayedFeedbackSnapshot,
} from "@/lib/delayed-feedback";
import type { ExperimentAlgorithm, ExperimentStatus, ExperimentSummary } from "@/lib/experiments";

const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

async function fetchJson<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, { cache: "no-store", ...init });
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

function probeClass(state: DelayedEndpointProbe["state"]) {
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

function formatAlgorithm(value: ExperimentAlgorithm | null) {
  if (!value) return "—";
  return value.replaceAll("_", " ");
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

function formatConversionValue(value: number, currency: string) {
  if (!currency && value === 0) return "—";

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

function conversionStatusClass(converted: boolean) {
  return converted ? "bg-green-100 text-green-800" : "bg-blue-100 text-blue-800";
}

interface LookupResponse<T> {
  ok: boolean;
  status: number;
  message: string;
  data?: T;
}

interface PendingRewardLookupResult extends LookupResponse<DelayedPendingReward> {}

interface UserPendingLookupResult extends LookupResponse<DelayedPendingRewardsByUser> {}

function mergeArmStats(experiment: ExperimentSummary | null, snapshot: DelayedFeedbackSnapshot | null) {
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

function isValidUuid(value: string) {
  return UUID_PATTERN.test(value.trim());
}

export function DelayedFeedbackPageClient({
  initialExperiments,
  initialSelectedExperimentId = null,
  initialSnapshot = null,
  loadFailed: initialLoadFailed = false,
}: {
  initialExperiments?: ExperimentSummary[];
  initialSelectedExperimentId?: string | null;
  initialSnapshot?: DelayedFeedbackSnapshot | null;
  loadFailed?: boolean;
}) {
  const hasInitialPayload = initialExperiments !== undefined;
  const t = useTranslations("feedback");
  const [experiments, setExperiments] = useState<ExperimentSummary[]>(initialExperiments ?? []);
  const [selectedId, setSelectedId] = useState(initialSelectedExperimentId ?? "");
  const [snapshot, setSnapshot] = useState<DelayedFeedbackSnapshot | null>(initialSnapshot ?? null);
  const [loadFailed, setLoadFailed] = useState(initialLoadFailed);
  const [isBootstrapping, setIsBootstrapping] = useState(!hasInitialPayload);
  const [isPending, startTransition] = useTransition();
  const [submission, setSubmission] = useState<DelayedConversionResult | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [pendingRewardId, setPendingRewardId] = useState("");
  const [pendingUserId, setPendingUserId] = useState("");
  const [pendingLookup, setPendingLookup] = useState<PendingRewardLookupResult | null>(null);
  const [userPendingLookup, setUserPendingLookup] = useState<UserPendingLookupResult | null>(null);
  const [isLookingUpPendingReward, setIsLookingUpPendingReward] = useState(false);
  const [isLookingUpUserPending, setIsLookingUpUserPending] = useState(false);
  const [form, setForm] = useState<DelayedConversionPayload>({
    transactionId: "",
    userId: "",
    conversionValue: 0,
    currency: "USD",
  });

  useEffect(() => {
    if (!isBootstrapping) return;

    startTransition(async () => {
      try {
        const data = await fetchJson<DelayedFeedbackDashboardData>("/api/admin/delayed-feedback/dashboard");
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
  const totalConversions = armRows.reduce((sum, arm) => sum + arm.conversions, 0);
  const totalRevenue = armRows.reduce((sum, arm) => sum + arm.revenue, 0);
  const formValid =
    isValidUuid(form.transactionId) &&
    isValidUuid(form.userId) &&
    Number.isFinite(form.conversionValue) &&
    form.conversionValue > 0 &&
    form.currency.trim().length > 0;
  const pendingRewardIdValid = isValidUuid(pendingRewardId);
  const pendingUserIdValid = isValidUuid(pendingUserId);

  const refreshSnapshot = (experimentId: string) => {
    setSelectedId(experimentId);
    setPendingLookup(null);
    setUserPendingLookup(null);
    startTransition(async () => {
      try {
        const data = await fetchJson<DelayedFeedbackSnapshot>(
          `/api/admin/delayed-feedback/snapshot?experimentId=${encodeURIComponent(experimentId)}`,
        );
        setSnapshot(data);
      } catch {
        toast.error(t("feedback.loadFailed"));
      }
    });
  };

  const lookupPendingReward = async () => {
    if (!pendingRewardIdValid) return;

    setIsLookingUpPendingReward(true);
    try {
      const res = await fetch(`/api/admin/delayed-feedback/pending/${encodeURIComponent(pendingRewardId.trim())}`, {
        cache: "no-store",
      });
      const body = await res.json().catch(() => ({}));
      const result: PendingRewardLookupResult = {
        ok: res.ok,
        status: res.status,
        message:
          (body as { message?: string; error?: string }).message ??
          (body as { error?: string }).error ??
          `HTTP ${res.status}`,
        data: res.ok ? (((body as { data?: DelayedPendingReward }).data ?? body) as DelayedPendingReward) : undefined,
      };

      setPendingLookup(result);
      if (!result.ok) {
        toast.error(result.message);
      }
    } catch {
      const result = {
        ok: false,
        status: 500,
        message: t("feedback.lookupFailed"),
      } satisfies PendingRewardLookupResult;
      setPendingLookup(result);
      toast.error(result.message);
    } finally {
      setIsLookingUpPendingReward(false);
    }
  };

  const lookupUserPendingRewards = async () => {
    if (!pendingUserIdValid) return;

    setIsLookingUpUserPending(true);
    try {
      const res = await fetch(
        `/api/admin/delayed-feedback/users/${encodeURIComponent(pendingUserId.trim())}/pending`,
        { cache: "no-store" },
      );
      const body = await res.json().catch(() => ({}));
      const result: UserPendingLookupResult = {
        ok: res.ok,
        status: res.status,
        message:
          (body as { message?: string; error?: string }).message ??
          (body as { error?: string }).error ??
          `HTTP ${res.status}`,
        data: res.ok
          ? (((body as { data?: DelayedPendingRewardsByUser }).data ?? body) as DelayedPendingRewardsByUser)
          : undefined,
      };

      setUserPendingLookup(result);
      if (!result.ok) {
        toast.error(result.message);
      }
    } catch {
      const result = {
        ok: false,
        status: 500,
        message: t("feedback.lookupFailed"),
      } satisfies UserPendingLookupResult;
      setUserPendingLookup(result);
      toast.error(result.message);
    } finally {
      setIsLookingUpUserPending(false);
    }
  };

  const submitConversion = async () => {
    if (!formValid) return;
    setIsSubmitting(true);
    try {
      const res = await fetch("/api/admin/delayed-feedback/conversions", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });
      const body = await res.json().catch(() => ({}));
      const result: DelayedConversionResult = {
        ok: res.ok,
        status: res.status,
        message:
          (body as { message?: string; error?: string }).message ??
          (body as { error?: string }).error ??
          `HTTP ${res.status}`,
        transactionId:
          (body as { transaction_id?: string }).transaction_id ??
          (body as { data?: { transaction_id?: string } }).data?.transaction_id,
      };
      setSubmission(result);
      if (result.ok) {
        toast.success(t("feedback.processed"));
      } else {
        toast.error(result.message);
      }
    } catch {
      const result = { ok: false, status: 500, message: t("feedback.submitFailed") } satisfies DelayedConversionResult;
      setSubmission(result);
      toast.error(result.message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const probeRows = snapshot
    ? [
        { key: "conversionIngest", probe: snapshot.probes.conversionIngest },
        { key: "pendingById", probe: snapshot.probes.pendingById },
        { key: "userPending", probe: snapshot.probes.userPending },
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
              { label: t("summary.conversions"), value: totalConversions.toLocaleString("en-US") },
              { label: t("summary.revenue"), value: formatRevenue(totalRevenue) },
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

            <Card>
              <CardHeader>
                <CardTitle className="text-sm">{t("process.title")}</CardTitle>
                <CardDescription>{t("process.description")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="grid gap-3 md:grid-cols-2">
                  <div className="space-y-1.5">
                    <p className="font-medium text-sm">{t("process.transactionId")}</p>
                    <Input
                      value={form.transactionId}
                      onChange={(event) => setForm((current) => ({ ...current, transactionId: event.target.value }))}
                      placeholder="11111111-1111-1111-1111-111111111111"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <p className="font-medium text-sm">{t("process.userId")}</p>
                    <Input
                      value={form.userId}
                      onChange={(event) => setForm((current) => ({ ...current, userId: event.target.value }))}
                      placeholder="22222222-2222-2222-2222-222222222222"
                    />
                  </div>
                </div>

                <div className="grid gap-3 md:grid-cols-2">
                  <div className="space-y-1.5">
                    <p className="font-medium text-sm">{t("process.value")}</p>
                    <Input
                      type="number"
                      min="0"
                      step="0.01"
                      value={String(form.conversionValue)}
                      onChange={(event) =>
                        setForm((current) => ({ ...current, conversionValue: Number(event.target.value) }))
                      }
                    />
                  </div>
                  <div className="space-y-1.5">
                    <p className="font-medium text-sm">{t("process.currency")}</p>
                    <Select
                      value={form.currency}
                      onValueChange={(value) => setForm((current) => ({ ...current, currency: value }))}
                    >
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="USD">USD</SelectItem>
                        <SelectItem value="EUR">EUR</SelectItem>
                        <SelectItem value="GBP">GBP</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                  <Activity className="mb-2 size-4" />
                  {t("process.warning")}
                </div>

                <Button size="sm" onClick={submitConversion} disabled={!formValid || isSubmitting}>
                  <Send className="size-4" />
                  {isSubmitting ? t("actions.processing") : t("actions.processConversion")}
                </Button>

                {submission ? (
                  <div className="rounded-md border p-3 text-sm">
                    <p className="font-medium">
                      {submission.ok ? t("process.resultSuccess") : t("process.resultFailure")} · HTTP{" "}
                      {submission.status}
                    </p>
                    <p className="mt-1 text-muted-foreground text-xs">{submission.message}</p>
                    {submission.transactionId ? (
                      <p className="mt-1 text-muted-foreground text-xs">{submission.transactionId}</p>
                    ) : null}
                  </div>
                ) : null}
              </CardContent>
            </Card>
          </div>

          <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">{t("lookup.pendingTitle")}</CardTitle>
                <CardDescription>{t("lookup.pendingDescription")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="space-y-1.5">
                  <p className="font-medium text-sm">{t("lookup.pendingId")}</p>
                  <Input
                    value={pendingRewardId}
                    onChange={(event) => setPendingRewardId(event.target.value)}
                    placeholder={t("lookup.pendingPlaceholder")}
                  />
                </div>

                <Button
                  size="sm"
                  onClick={lookupPendingReward}
                  disabled={!pendingRewardIdValid || isLookingUpPendingReward}
                >
                  <Activity className="size-4" />
                  {isLookingUpPendingReward ? t("lookup.lookingUp") : t("lookup.lookupPending")}
                </Button>

                {pendingLookup ? (
                  <div className="rounded-md border p-3 text-sm">
                    <p className="font-medium">{t("lookup.httpStatus", { status: pendingLookup.status })}</p>
                    <p className={`mt-1 text-xs ${pendingLookup.ok ? "text-muted-foreground" : "text-destructive"}`}>
                      {pendingLookup.message}
                    </p>

                    {pendingLookup.ok && pendingLookup.data ? (
                      <div className="mt-3 grid gap-3 md:grid-cols-2">
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.rewardId")}</p>
                          <p className="break-all font-medium">{pendingLookup.data.ID}</p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.statusLabel")}</p>
                          <Badge className={conversionStatusClass(pendingLookup.data.Converted)}>
                            {pendingLookup.data.Converted ? t("lookup.converted") : t("lookup.pending")}
                          </Badge>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.experimentId")}</p>
                          <p className="break-all font-medium">{pendingLookup.data.ExperimentID}</p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.armId")}</p>
                          <p className="break-all font-medium">{pendingLookup.data.ArmID}</p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.userId")}</p>
                          <p className="break-all font-medium">{pendingLookup.data.UserID}</p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.conversionValue")}</p>
                          <p className="font-medium">
                            {formatConversionValue(
                              pendingLookup.data.ConversionValue,
                              pendingLookup.data.ConversionCurrency,
                            )}
                          </p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.assignedAt")}</p>
                          <p className="font-medium">{formatDateTime(pendingLookup.data.AssignedAt)}</p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.expiresAt")}</p>
                          <p className="font-medium">{formatDateTime(pendingLookup.data.ExpiresAt)}</p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.convertedAt")}</p>
                          <p className="font-medium">{formatDateTime(pendingLookup.data.ConvertedAt)}</p>
                        </div>
                        <div>
                          <p className="text-muted-foreground text-xs">{t("lookup.processedAt")}</p>
                          <p className="font-medium">{formatDateTime(pendingLookup.data.ProcessedAt)}</p>
                        </div>
                      </div>
                    ) : null}
                  </div>
                ) : null}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-sm">{t("lookup.userTitle")}</CardTitle>
                <CardDescription>{t("lookup.userDescription")}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="space-y-1.5">
                  <p className="font-medium text-sm">{t("lookup.userId")}</p>
                  <Input
                    value={pendingUserId}
                    onChange={(event) => setPendingUserId(event.target.value)}
                    placeholder={t("lookup.userPlaceholder")}
                  />
                </div>

                <Button
                  size="sm"
                  onClick={lookupUserPendingRewards}
                  disabled={!pendingUserIdValid || isLookingUpUserPending}
                >
                  <Activity className="size-4" />
                  {isLookingUpUserPending ? t("lookup.lookingUp") : t("lookup.lookupUser")}
                </Button>

                {userPendingLookup ? (
                  <div className="rounded-md border p-3 text-sm">
                    <p className="font-medium">{t("lookup.httpStatus", { status: userPendingLookup.status })}</p>
                    <p className={`mt-1 text-xs ${userPendingLookup.ok ? "text-muted-foreground" : "text-destructive"}`}>
                      {userPendingLookup.message}
                    </p>

                    {userPendingLookup.ok && userPendingLookup.data ? (
                      <>
                        <p className="mt-3 text-muted-foreground text-xs">{t("lookup.userResultDescription")}</p>
                        <p className="mt-1 break-all font-medium">{userPendingLookup.data.user_id}</p>

                        {userPendingLookup.data.rewards.length === 0 ? (
                          <p className="mt-3 text-muted-foreground text-xs">{t("lookup.noRewards")}</p>
                        ) : (
                          <div className="mt-3 overflow-x-auto">
                            <Table>
                              <TableHeader>
                                <TableRow>
                                  <TableHead>{t("lookup.rewardId")}</TableHead>
                                  <TableHead>{t("lookup.experimentId")}</TableHead>
                                  <TableHead>{t("lookup.armId")}</TableHead>
                                  <TableHead>{t("lookup.assignedAt")}</TableHead>
                                  <TableHead>{t("lookup.expiresAt")}</TableHead>
                                  <TableHead>{t("lookup.statusLabel")}</TableHead>
                                  <TableHead>{t("lookup.conversionValue")}</TableHead>
                                </TableRow>
                              </TableHeader>
                              <TableBody>
                                {userPendingLookup.data.rewards.map((reward) => (
                                  <TableRow key={reward.ID}>
                                    <TableCell className="max-w-44 break-all text-xs">{reward.ID}</TableCell>
                                    <TableCell className="max-w-44 break-all text-xs">{reward.ExperimentID}</TableCell>
                                    <TableCell className="max-w-44 break-all text-xs">{reward.ArmID}</TableCell>
                                    <TableCell className="text-xs">{formatDateTime(reward.AssignedAt)}</TableCell>
                                    <TableCell className="text-xs">{formatDateTime(reward.ExpiresAt)}</TableCell>
                                    <TableCell>
                                      <Badge className={conversionStatusClass(reward.Converted)}>
                                        {reward.Converted ? t("lookup.converted") : t("lookup.pending")}
                                      </Badge>
                                    </TableCell>
                                    <TableCell className="text-xs">
                                      {formatConversionValue(reward.ConversionValue, reward.ConversionCurrency)}
                                    </TableCell>
                                  </TableRow>
                                ))}
                              </TableBody>
                            </Table>
                          </div>
                        )}
                      </>
                    ) : null}
                  </div>
                ) : null}
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
              <CardTitle className="text-sm">{t("notes.title")}</CardTitle>
              <CardDescription>{t("notes.description")}</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-3 text-sm md:grid-cols-3">
              <div className="rounded-md border p-3">
                <p className="font-medium">{t("notes.pendingTitle")}</p>
                <p className="mt-1 text-muted-foreground text-xs">{t("notes.pendingBody")}</p>
              </div>
              <div className="rounded-md border p-3">
                <p className="font-medium">{t("notes.conversionTitle")}</p>
                <p className="mt-1 text-muted-foreground text-xs">{t("notes.conversionBody")}</p>
              </div>
              <div className="rounded-md border p-3">
                <p className="font-medium">{t("notes.operatorTitle")}</p>
                <p className="mt-1 text-muted-foreground text-xs">{t("notes.operatorBody")}</p>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
