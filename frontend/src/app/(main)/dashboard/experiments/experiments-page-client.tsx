"use client";

import { Fragment, useEffect, useState } from "react";

import { useRouter } from "next/navigation";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  completeExperimentAction,
  createExperimentAction,
  pauseExperimentAction,
  resumeExperimentAction,
  updateExperimentAction,
} from "@/actions/experiments";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import {
  EMPTY_EXPERIMENT_INPUT,
  type ExperimentAlgorithm,
  type ExperimentInput,
  type ExperimentStatus,
  type ExperimentSummary,
  formatExperimentLifecycleCode,
  getExperimentLifecycleReason,
  getExperimentLifecycleReasonKey,
  getExperimentLifecycleSourceKey,
} from "@/lib/experiments";
import { formatAdminDateTime, toDateTimeLocalInputValue } from "@/lib/time";

const formSchema = z
  .object({
    name: z.string().trim().min(1),
    description: z.string(),
    status: z.enum(["draft", "running", "paused", "completed"]),
    algorithm_type: z.enum(["thompson_sampling", "ucb", "epsilon_greedy"]),
    is_bandit: z.boolean(),
    min_sample_size: z.string().trim(),
    confidence_threshold_percent: z.string().trim(),
    start_at: z.string(),
    end_at: z.string(),
    control_name: z.string().trim().min(1),
    control_description: z.string(),
    control_weight: z.string().trim(),
    variant_name: z.string().trim().min(1),
    variant_description: z.string(),
    variant_weight: z.string().trim(),
  })
  .superRefine((value, ctx) => {
    for (const [field, raw] of [
      ["min_sample_size", value.min_sample_size],
      ["confidence_threshold_percent", value.confidence_threshold_percent],
      ["control_weight", value.control_weight],
      ["variant_weight", value.variant_weight],
    ] as const) {
      const parsed = Number(raw);
      if (!Number.isFinite(parsed) || parsed <= 0) {
        ctx.addIssue({ code: "custom", message: "Must be greater than zero", path: [field] });
      }
    }

    const threshold = Number(value.confidence_threshold_percent);
    if (Number.isFinite(threshold) && threshold > 100) {
      ctx.addIssue({ code: "custom", message: "Cannot exceed 100", path: ["confidence_threshold_percent"] });
    }

    if (value.start_at && value.end_at && value.end_at < value.start_at) {
      ctx.addIssue({ code: "custom", message: "End must be after start", path: ["end_at"] });
    }
  });

type ExperimentFormValues = z.infer<typeof formSchema>;

type ExperimentEditorValidationCode = "name" | "minSampleSize" | "confidenceThreshold" | "confidenceRange" | "endAt";

type ExperimentEditorFormValues = {
  name: string;
  description: string;
  algorithm_type: ExperimentAlgorithm;
  is_bandit: boolean;
  min_sample_size: string;
  confidence_threshold_percent: string;
  start_at: string;
  end_at: string;
};

const EMPTY_FORM_VALUES: ExperimentFormValues = {
  name: EMPTY_EXPERIMENT_INPUT.name,
  description: EMPTY_EXPERIMENT_INPUT.description,
  status: EMPTY_EXPERIMENT_INPUT.status,
  algorithm_type: EMPTY_EXPERIMENT_INPUT.algorithm_type,
  is_bandit: EMPTY_EXPERIMENT_INPUT.is_bandit,
  min_sample_size: EMPTY_EXPERIMENT_INPUT.min_sample_size.toString(),
  confidence_threshold_percent: EMPTY_EXPERIMENT_INPUT.confidence_threshold_percent.toString(),
  start_at: "",
  end_at: "",
  control_name: EMPTY_EXPERIMENT_INPUT.arms[0].name,
  control_description: EMPTY_EXPERIMENT_INPUT.arms[0].description,
  control_weight: EMPTY_EXPERIMENT_INPUT.arms[0].traffic_weight.toString(),
  variant_name: EMPTY_EXPERIMENT_INPUT.arms[1].name,
  variant_description: EMPTY_EXPERIMENT_INPUT.arms[1].description,
  variant_weight: EMPTY_EXPERIMENT_INPUT.arms[1].traffic_weight.toString(),
};

function fieldError(error?: { message?: string }) {
  if (!error?.message) return null;
  return <p className="text-destructive text-xs">{error.message}</p>;
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

function formatLifecycleReason(t: ReturnType<typeof useTranslations>, experiment: ExperimentSummary) {
  const reason = getExperimentLifecycleReason(experiment.latest_lifecycle_audit);
  const reasonKey = getExperimentLifecycleReasonKey(reason);
  return reasonKey ? t(`table.${reasonKey}`) : formatExperimentLifecycleCode(reason);
}

function formatLifecycleSource(t: ReturnType<typeof useTranslations>, experiment: ExperimentSummary) {
  const sourceKey = getExperimentLifecycleSourceKey(experiment.latest_lifecycle_audit?.source);
  return sourceKey ? t(`table.${sourceKey}`) : formatExperimentLifecycleCode(experiment.latest_lifecycle_audit?.source);
}

function buildEditorValues(experiment: ExperimentSummary): ExperimentEditorFormValues {
  return {
    name: experiment.name,
    description: experiment.description,
    algorithm_type: experiment.algorithm_type ?? "thompson_sampling",
    is_bandit: experiment.is_bandit,
    min_sample_size: experiment.min_sample_size.toString(),
    confidence_threshold_percent: experiment.confidence_threshold_percent.toString(),
    start_at: toDateTimeLocalInputValue(experiment.start_at),
    end_at: toDateTimeLocalInputValue(experiment.end_at),
  };
}

function validateEditorValues(values: ExperimentEditorFormValues): ExperimentEditorValidationCode | null {
  if (!values.name.trim()) return "name";

  const minSampleSize = Number(values.min_sample_size);
  if (!Number.isFinite(minSampleSize) || minSampleSize <= 0) return "minSampleSize";

  const threshold = Number(values.confidence_threshold_percent);
  if (!Number.isFinite(threshold) || threshold <= 0) return "confidenceThreshold";
  if (threshold > 100) return "confidenceRange";

  if (values.start_at && values.end_at && values.end_at < values.start_at) return "endAt";

  return null;
}

function toPayload(values: ExperimentFormValues): ExperimentInput {
  return {
    name: values.name.trim(),
    description: values.description.trim(),
    status: values.status,
    algorithm_type: values.algorithm_type,
    is_bandit: values.is_bandit,
    min_sample_size: Number(values.min_sample_size),
    confidence_threshold_percent: Number(values.confidence_threshold_percent),
    start_at: values.start_at ? new Date(values.start_at).toISOString() : null,
    end_at: values.end_at ? new Date(values.end_at).toISOString() : null,
    arms: [
      {
        name: values.control_name.trim(),
        description: values.control_description.trim(),
        is_control: true,
        traffic_weight: Number(values.control_weight),
      },
      {
        name: values.variant_name.trim(),
        description: values.variant_description.trim(),
        is_control: false,
        traffic_weight: Number(values.variant_weight),
      },
    ],
  };
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

export function ExperimentsPageClient({
  initialExperiments,
  loadFailed,
}: {
  initialExperiments: ExperimentSummary[];
  loadFailed: boolean;
}) {
  const t = useTranslations("experiments");
  const router = useRouter();
  const [hasHydrated, setHasHydrated] = useState(false);
  const [experiments, setExperiments] = useState(initialExperiments);
  const [pendingCreate, setPendingCreate] = useState(false);
  const [pendingAction, setPendingAction] = useState<string | null>(null);
  const [editingExperimentId, setEditingExperimentId] = useState<string | null>(null);
  const [editorValues, setEditorValues] = useState<ExperimentEditorFormValues | null>(null);
  const [pendingEditId, setPendingEditId] = useState<string | null>(null);
  const form = useForm<ExperimentFormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: EMPTY_FORM_VALUES,
  });

  useEffect(() => {
    setHasHydrated(true);
  }, []);

  const runningCount = experiments.filter((experiment) => experiment.status === "running").length;
  const draftCount = experiments.filter((experiment) => experiment.status === "draft").length;
  const totalRevenue = experiments.reduce((sum, experiment) => sum + experiment.total_revenue, 0);

  function syncExperiment(updatedExperiment: ExperimentSummary) {
    setExperiments((current) => current.map((item) => (item.id === updatedExperiment.id ? updatedExperiment : item)));
  }

  function startEditingExperiment(experiment: ExperimentSummary) {
    if (experiment.status !== "draft") return;
    setEditingExperimentId(experiment.id);
    setEditorValues(buildEditorValues(experiment));
  }

  function cancelEditingExperiment() {
    setEditingExperimentId(null);
    setEditorValues(null);
  }

  function updateEditorValue<Key extends keyof ExperimentEditorFormValues>(
    key: Key,
    value: ExperimentEditorFormValues[Key],
  ) {
    setEditorValues((current) => (current ? { ...current, [key]: value } : current));
  }

  const createExperiment = form.handleSubmit(async (values) => {
    setPendingCreate(true);
    const result = await createExperimentAction(toPayload(values));
    setPendingCreate(false);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.createFailed"));
      return;
    }

    setExperiments((current) => [result.data, ...current]);
    form.reset(EMPTY_FORM_VALUES);
    toast.success(t("feedback.experimentCreated"));
    router.refresh();
  });

  async function runLifecycleAction(experiment: ExperimentSummary, action: "pause" | "resume" | "complete") {
    setPendingAction(`${action}:${experiment.id}`);
    const result =
      action === "pause"
        ? await pauseExperimentAction(experiment.id)
        : action === "complete"
          ? await completeExperimentAction(experiment.id)
          : await resumeExperimentAction(experiment.id);
    setPendingAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.statusFailed"));
      return;
    }

    if (editingExperimentId === result.data.id && result.data.status !== "draft") {
      cancelEditingExperiment();
    }

    syncExperiment(result.data);
    toast.success(t("feedback.statusUpdated"));
    router.refresh();
  }

  async function saveExperimentMetadata(experiment: ExperimentSummary) {
    if (!editorValues || editingExperimentId !== experiment.id) return;

    const validationError = validateEditorValues(editorValues);
    if (validationError) {
      toast.error(t(`editor.validation.${validationError}`));
      return;
    }

    setPendingEditId(experiment.id);
    const result = await updateExperimentAction(experiment.id, {
      name: editorValues.name.trim(),
      description: editorValues.description.trim(),
      algorithm_type: editorValues.algorithm_type,
      is_bandit: editorValues.is_bandit,
      min_sample_size: Number(editorValues.min_sample_size),
      confidence_threshold_percent: Number(editorValues.confidence_threshold_percent),
      start_at: editorValues.start_at ? new Date(editorValues.start_at).toISOString() : null,
      end_at: editorValues.end_at ? new Date(editorValues.end_at).toISOString() : null,
    });
    setPendingEditId(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.updateFailed"));
      return;
    }

    syncExperiment(result.data);
    cancelEditingExperiment();
    toast.success(t("feedback.experimentUpdated"));
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="font-semibold text-2xl tracking-tight">{t("title")}</h1>
        <p className="mt-0.5 text-muted-foreground text-sm">{t("subtitle")}</p>
      </div>

      {loadFailed ? (
        <Card className="border-destructive/40">
          <CardContent className="pt-6 text-destructive text-sm">{t("states.loadFailed")}</CardContent>
        </Card>
      ) : null}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-4">
        {[
          { label: t("summary.total"), value: experiments.length },
          { label: t("summary.running"), value: runningCount },
          { label: t("summary.draft"), value: draftCount },
          { label: t("summary.revenue"), value: formatRevenue(totalRevenue) },
        ].map((item) => (
          <Card key={item.label}>
            <CardContent className="pt-6">
              <p className="font-semibold text-muted-foreground text-xs uppercase tracking-widest">{item.label}</p>
              <p className="mt-2 font-bold text-2xl tabular-nums">{item.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("table.title")}</CardTitle>
          <CardDescription>{t("table.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          {experiments.length === 0 ? (
            <div className="py-12 text-center text-muted-foreground text-sm">{t("states.empty")}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("table.name")}</TableHead>
                  <TableHead>{t("table.status")}</TableHead>
                  <TableHead>{t("table.algorithm")}</TableHead>
                  <TableHead>{t("table.arms")}</TableHead>
                  <TableHead>{t("table.samples")}</TableHead>
                  <TableHead>{t("table.conversions")}</TableHead>
                  <TableHead>{t("table.assignments")}</TableHead>
                  <TableHead>{t("table.revenue")}</TableHead>
                  <TableHead>{t("table.confidence")}</TableHead>
                  <TableHead>{t("table.updated")}</TableHead>
                  <TableHead>{t("table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {experiments.map((experiment) => {
                  const actionKey = pendingAction ?? "";
                  const isRowPending = actionKey.endsWith(`:${experiment.id}`);
                  const isEditingCurrentRow = editingExperimentId === experiment.id && editorValues !== null;
                  const isEditDirty = isEditingCurrentRow
                    ? JSON.stringify(editorValues) !== JSON.stringify(buildEditorValues(experiment))
                    : false;
                  const isRowBusy = isRowPending || pendingEditId === experiment.id;
                  const actions: Array<{ key: "pause" | "resume" | "complete"; label: string }> =
                    experiment.status === "draft"
                      ? [{ key: "resume", label: t("actions.start") }]
                      : experiment.status === "running"
                        ? [
                            { key: "pause", label: t("actions.pause") },
                            { key: "complete", label: t("actions.complete") },
                          ]
                        : experiment.status === "paused"
                          ? [
                              { key: "resume", label: t("actions.resume") },
                              { key: "complete", label: t("actions.complete") },
                            ]
                          : [];

                  return (
                    <Fragment key={experiment.id}>
                      <TableRow>
                        <TableCell className="min-w-72 whitespace-normal align-top">
                          <div className="max-w-sm">
                            <p className="break-words font-medium">{experiment.name}</p>
                            <p className="whitespace-normal break-words text-muted-foreground text-xs">
                              {experiment.description || "—"}
                            </p>
                            {experiment.latest_lifecycle_audit ? (
                              <div className="mt-2 rounded-md border border-dashed p-2 text-[11px] text-muted-foreground">
                                <p>
                                  {t("table.latestLifecycleLabel")}: {formatLifecycleSource(t, experiment)} ·{" "}
                                  {formatLifecycleReason(t, experiment)}
                                </p>
                                <p className="mt-1">
                                  {t(`status.${experiment.latest_lifecycle_audit.from_status}`)} →{" "}
                                  {t(`status.${experiment.latest_lifecycle_audit.to_status}`)} ·{" "}
                                  {hasHydrated
                                    ? formatAdminDateTime(experiment.latest_lifecycle_audit.created_at)
                                    : "—"}
                                </p>
                              </div>
                            ) : null}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge className={statusClass(experiment.status)}>{t(`status.${experiment.status}`)}</Badge>
                        </TableCell>
                        <TableCell>
                          <div className="text-sm">
                            <p>{formatAlgorithm(experiment.algorithm_type)}</p>
                            <p className="text-muted-foreground text-xs">
                              {experiment.is_bandit ? t("table.banditEnabled") : t("table.banditDisabled")}
                            </p>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col gap-1 text-xs">
                            {experiment.arms.map((arm) => (
                              <span key={arm.id}>
                                {arm.is_control ? t("table.controlArm") : t("table.variantArm")} {arm.name}
                              </span>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell className="font-mono text-sm">{experiment.total_samples}</TableCell>
                        <TableCell className="font-mono text-sm">{experiment.total_conversions}</TableCell>
                        <TableCell className="font-mono text-sm">{experiment.active_assignments}</TableCell>
                        <TableCell className="font-mono text-sm">{formatRevenue(experiment.total_revenue)}</TableCell>
                        <TableCell className="text-xs">
                          <div className="flex flex-col gap-1">
                            <span>{experiment.confidence_threshold_percent.toFixed(0)}%</span>
                            <span className="text-muted-foreground">
                              {experiment.winner_confidence_percent === null
                                ? t("table.noWinner")
                                : `${experiment.winner_confidence_percent.toFixed(1)}%`}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell className="text-muted-foreground text-xs">
                          {hasHydrated ? formatAdminDateTime(experiment.updated_at) : "—"}
                        </TableCell>
                        <TableCell>
                          {actions.length === 0 && experiment.status !== "draft" ? (
                            <span className="text-muted-foreground text-xs">—</span>
                          ) : (
                            <div className="flex flex-wrap gap-2">
                              {actions.map((action) => {
                                const currentActionKey = `${action.key}:${experiment.id}`;
                                return (
                                  <Button
                                    key={currentActionKey}
                                    variant="outline"
                                    size="sm"
                                    disabled={isRowBusy}
                                    onClick={() => void runLifecycleAction(experiment, action.key)}
                                  >
                                    {pendingAction === currentActionKey ? t("feedback.updating") : action.label}
                                  </Button>
                                );
                              })}
                              {experiment.status === "draft" && !isEditingCurrentRow ? (
                                <Button
                                  variant="outline"
                                  size="sm"
                                  disabled={isRowBusy}
                                  onClick={() => startEditingExperiment(experiment)}
                                >
                                  {t("actions.edit")}
                                </Button>
                              ) : null}
                            </div>
                          )}
                        </TableCell>
                      </TableRow>
                      {isEditingCurrentRow ? (
                        <TableRow>
                          <TableCell colSpan={11} className="bg-muted/20">
                            <div className="rounded-md border border-dashed p-4">
                              <div>
                                <p className="font-medium text-sm">{t("editor.title")}</p>
                                <p className="mt-1 text-muted-foreground text-xs">{t("editor.description")}</p>
                              </div>

                              <form
                                className="mt-4 space-y-4"
                                onSubmit={(event) => {
                                  event.preventDefault();
                                  void saveExperimentMetadata(experiment);
                                }}
                              >
                                <div className="grid gap-4 md:grid-cols-2">
                                  <div className="space-y-1">
                                    <p className="font-medium text-xs">{t("form.name")}</p>
                                    <Input
                                      value={editorValues.name}
                                      onChange={(event) => updateEditorValue("name", event.target.value)}
                                    />
                                  </div>

                                  <div className="space-y-1">
                                    <p className="font-medium text-xs">{t("form.algorithm")}</p>
                                    <Select
                                      value={editorValues.algorithm_type}
                                      onValueChange={(value) =>
                                        updateEditorValue("algorithm_type", value as ExperimentAlgorithm)
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
                                    <p className="font-medium text-xs">{t("form.descriptionLabel")}</p>
                                    <Textarea
                                      rows={3}
                                      value={editorValues.description}
                                      onChange={(event) => updateEditorValue("description", event.target.value)}
                                    />
                                  </div>

                                  <div className="space-y-1">
                                    <p className="font-medium text-xs">{t("form.minSampleSize")}</p>
                                    <Input
                                      inputMode="numeric"
                                      value={editorValues.min_sample_size}
                                      onChange={(event) => updateEditorValue("min_sample_size", event.target.value)}
                                    />
                                  </div>

                                  <div className="space-y-1">
                                    <p className="font-medium text-xs">{t("form.confidenceThreshold")}</p>
                                    <Input
                                      inputMode="decimal"
                                      value={editorValues.confidence_threshold_percent}
                                      onChange={(event) =>
                                        updateEditorValue("confidence_threshold_percent", event.target.value)
                                      }
                                    />
                                  </div>

                                  <div className="space-y-1">
                                    <p className="font-medium text-xs">{t("form.startAt")}</p>
                                    <Input
                                      type="datetime-local"
                                      value={hasHydrated ? editorValues.start_at : ""}
                                      onChange={(event) => updateEditorValue("start_at", event.target.value)}
                                    />
                                  </div>

                                  <div className="space-y-1">
                                    <p className="font-medium text-xs">{t("form.endAt")}</p>
                                    <Input
                                      type="datetime-local"
                                      value={hasHydrated ? editorValues.end_at : ""}
                                      onChange={(event) => updateEditorValue("end_at", event.target.value)}
                                    />
                                  </div>
                                </div>

                                <div className="flex items-center gap-2">
                                  <Switch
                                    checked={editorValues.is_bandit}
                                    onCheckedChange={(checked) => updateEditorValue("is_bandit", checked)}
                                  />
                                  <span className="text-sm">{t("form.isBandit")}</span>
                                </div>

                                <div className="rounded-md border border-dashed p-3 text-muted-foreground text-xs">
                                  {t("editor.armHint")}
                                </div>

                                <div className="flex flex-wrap gap-2">
                                  <Button
                                    type="submit"
                                    size="sm"
                                    disabled={pendingEditId === experiment.id || !isEditDirty}
                                  >
                                    {pendingEditId === experiment.id ? t("feedback.updating") : t("actions.save")}
                                  </Button>
                                  <Button
                                    type="button"
                                    size="sm"
                                    variant="outline"
                                    disabled={pendingEditId === experiment.id}
                                    onClick={cancelEditingExperiment}
                                  >
                                    {t("actions.cancel")}
                                  </Button>
                                </div>
                              </form>
                            </div>
                          </TableCell>
                        </TableRow>
                      ) : null}
                    </Fragment>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("form.title")}</CardTitle>
          <CardDescription>{t("form.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={createExperiment}>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.name")}</p>
                <Input placeholder={t("form.namePlaceholder")} {...form.register("name")} />
                {fieldError(form.formState.errors.name)}
              </div>

              <Controller
                control={form.control}
                name="status"
                render={({ field }) => (
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.status")}</p>
                    <Select value={field.value} onValueChange={(value) => field.onChange(value as ExperimentStatus)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="draft">{t("status.draft")}</SelectItem>
                        <SelectItem value="running">{t("status.running")}</SelectItem>
                        <SelectItem value="paused">{t("status.paused")}</SelectItem>
                        <SelectItem value="completed">{t("status.completed")}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                )}
              />

              <div className="space-y-1 md:col-span-2">
                <p className="font-medium text-xs">{t("form.descriptionLabel")}</p>
                <Textarea placeholder={t("form.descriptionPlaceholder")} rows={3} {...form.register("description")} />
              </div>

              <Controller
                control={form.control}
                name="algorithm_type"
                render={({ field }) => (
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.algorithm")}</p>
                    <Select value={field.value} onValueChange={(value) => field.onChange(value as ExperimentAlgorithm)}>
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
                )}
              />

              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.minSampleSize")}</p>
                <Input inputMode="numeric" placeholder="100" {...form.register("min_sample_size")} />
                {fieldError(form.formState.errors.min_sample_size)}
              </div>

              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.confidenceThreshold")}</p>
                <Input inputMode="decimal" placeholder="95" {...form.register("confidence_threshold_percent")} />
                {fieldError(form.formState.errors.confidence_threshold_percent)}
              </div>

              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.startAt")}</p>
                <Input type="datetime-local" {...form.register("start_at")} />
              </div>

              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.endAt")}</p>
                <Input type="datetime-local" {...form.register("end_at")} />
                {fieldError(form.formState.errors.end_at)}
              </div>
            </div>

            <Controller
              control={form.control}
              name="is_bandit"
              render={({ field }) => (
                <div className="flex items-center gap-2">
                  <Switch checked={field.value} onCheckedChange={field.onChange} />
                  <span className="text-sm">{t("form.isBandit")}</span>
                </div>
              )}
            />

            <div className="grid gap-4 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("form.controlTitle")}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.armName")}</p>
                    <Input {...form.register("control_name")} />
                    {fieldError(form.formState.errors.control_name)}
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.armDescription")}</p>
                    <Textarea rows={2} {...form.register("control_description")} />
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.trafficWeight")}</p>
                    <Input inputMode="decimal" {...form.register("control_weight")} />
                    {fieldError(form.formState.errors.control_weight)}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-sm">{t("form.variantTitle")}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.armName")}</p>
                    <Input {...form.register("variant_name")} />
                    {fieldError(form.formState.errors.variant_name)}
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.armDescription")}</p>
                    <Textarea rows={2} {...form.register("variant_description")} />
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.trafficWeight")}</p>
                    <Input inputMode="decimal" {...form.register("variant_weight")} />
                    {fieldError(form.formState.errors.variant_weight)}
                  </div>
                </CardContent>
              </Card>
            </div>

            <div className="flex gap-2">
              <Button type="submit" size="sm" disabled={pendingCreate}>
                {pendingCreate ? t("feedback.creating") : t("actions.create")}
              </Button>
              <Button type="button" size="sm" variant="outline" onClick={() => form.reset(EMPTY_FORM_VALUES)}>
                {t("actions.reset")}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
