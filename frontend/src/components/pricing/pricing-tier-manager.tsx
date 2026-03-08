/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-08 09:10
 * Last Updated: 2026-03-08 09:17
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

"use client";

import { useCallback, useEffect, useState } from "react";

import { useRouter } from "next/navigation";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  activatePricingTierAction,
  createPricingTierAction,
  deactivatePricingTierAction,
  updatePricingTierAction,
} from "@/actions/pricing";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { EMPTY_PRICING_TIER_INPUT, type PricingTier, type PricingTierInput } from "@/lib/pricing-tiers";

const formSchema = z
  .object({
    name: z.string().trim().min(1),
    description: z.string(),
    monthly_price: z.string().trim(),
    annual_price: z.string().trim(),
    currency: z
      .string()
      .trim()
      .regex(/^[A-Za-z]{3}$/),
    features: z.string(),
    is_active: z.boolean(),
  })
  .superRefine((value, ctx) => {
    const monthly = value.monthly_price.trim();
    const annual = value.annual_price.trim();

    if (!monthly && !annual) {
      ctx.addIssue({ code: "custom", message: "Provide at least one price", path: ["monthly_price"] });
      ctx.addIssue({ code: "custom", message: "Provide at least one price", path: ["annual_price"] });
    }

    for (const [field, raw] of [
      ["monthly_price", monthly],
      ["annual_price", annual],
    ] as const) {
      if (!raw) continue;
      const parsed = Number(raw);
      if (!Number.isFinite(parsed) || parsed <= 0) {
        ctx.addIssue({ code: "custom", message: "Price must be greater than zero", path: [field] });
      }
    }
  });

type PricingFormValues = z.infer<typeof formSchema>;

const EMPTY_FORM_VALUES: PricingFormValues = {
  name: EMPTY_PRICING_TIER_INPUT.name,
  description: EMPTY_PRICING_TIER_INPUT.description,
  monthly_price: "",
  annual_price: "",
  currency: EMPTY_PRICING_TIER_INPUT.currency,
  features: "",
  is_active: EMPTY_PRICING_TIER_INPUT.is_active,
};

function fieldError(error?: { message?: string }) {
  if (!error?.message) return null;
  return <p className="text-destructive text-xs">{error.message}</p>;
}

function formatMoney(value: number | null, currency: string) {
  if (value === null) return "—";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function tierToFormValues(tier: PricingTier): PricingFormValues {
  return {
    name: tier.name,
    description: tier.description,
    monthly_price: tier.monthly_price?.toString() ?? "",
    annual_price: tier.annual_price?.toString() ?? "",
    currency: tier.currency,
    features: tier.features.join("\n"),
    is_active: tier.is_active,
  };
}

function toPayload(values: PricingFormValues): PricingTierInput {
  const parsePrice = (raw: string) => {
    const trimmed = raw.trim();
    return trimmed ? Number(trimmed) : null;
  };

  return {
    name: values.name.trim(),
    description: values.description.trim(),
    monthly_price: parsePrice(values.monthly_price),
    annual_price: parsePrice(values.annual_price),
    currency: values.currency.trim().toUpperCase(),
    features: values.features
      .split("\n")
      .map((item) => item.trim())
      .filter(Boolean),
    is_active: values.is_active,
  };
}

export function PricingTierManager({
  initialTiers,
  loadFailed,
  embedded = false,
}: {
  initialTiers: PricingTier[];
  loadFailed: boolean;
  embedded?: boolean;
}) {
  const t = useTranslations("pricing");
  const router = useRouter();
  const [hasHydrated, setHasHydrated] = useState(false);
  const [tiers, setTiers] = useState(initialTiers);
  const [editingTierId, setEditingTierId] = useState<string | null>(initialTiers[0]?.id ?? null);
  const [pendingAction, setPendingAction] = useState<string | null>(null);
  const form = useForm<PricingFormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: initialTiers[0] ? tierToFormValues(initialTiers[0]) : EMPTY_FORM_VALUES,
  });

  const applyFormValues = useCallback(
    (values: PricingFormValues) => {
      form.reset(values);
      form.setValue("name", values.name);
      form.setValue("description", values.description);
      form.setValue("monthly_price", values.monthly_price);
      form.setValue("annual_price", values.annual_price);
      form.setValue("currency", values.currency);
      form.setValue("features", values.features);
      form.setValue("is_active", values.is_active);
    },
    [form],
  );

  useEffect(() => {
    setHasHydrated(true);
  }, []);

  useEffect(() => {
    setTiers(initialTiers);

    if (initialTiers.length === 0) {
      if (editingTierId !== null) {
        setEditingTierId(null);
      }
      applyFormValues(EMPTY_FORM_VALUES);
      return;
    }

    if (editingTierId === null) {
      applyFormValues(EMPTY_FORM_VALUES);
      return;
    }

    const nextTier = initialTiers.find((tier) => tier.id === editingTierId) ?? initialTiers[0];

    if (editingTierId !== nextTier.id) {
      setEditingTierId(nextTier.id);
    }

    applyFormValues(tierToFormValues(nextTier));
  }, [applyFormValues, editingTierId, initialTiers]);

  const activeCount = tiers.filter((tier) => tier.is_active).length;

  function resetToNewTier() {
    setEditingTierId(null);
    applyFormValues(EMPTY_FORM_VALUES);
  }

  function selectTier(tier: PricingTier) {
    setEditingTierId(tier.id);
    applyFormValues(tierToFormValues(tier));
  }

  const saveTier = form.handleSubmit(async (values) => {
    setPendingAction("save");
    const result = editingTierId
      ? await updatePricingTierAction(editingTierId, toPayload(values))
      : await createPricingTierAction(toPayload(values));
    setPendingAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.saveFailed"));
      return;
    }

    const nextTier = result.data;
    setTiers((current) =>
      editingTierId ? current.map((tier) => (tier.id === nextTier.id ? nextTier : tier)) : [nextTier, ...current],
    );
    setEditingTierId(nextTier.id);
    applyFormValues(tierToFormValues(nextTier));
    toast.success(editingTierId ? t("feedback.tierUpdated") : t("feedback.tierCreated"));
    router.refresh();
  });

  async function toggleTier(tier: PricingTier) {
    setPendingAction(`toggle:${tier.id}`);
    const result = tier.is_active
      ? await deactivatePricingTierAction(tier.id)
      : await activatePricingTierAction(tier.id);
    setPendingAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.statusFailed"));
      return;
    }

    const updatedTier = result.data;
    setTiers((current) => current.map((item) => (item.id === updatedTier.id ? updatedTier : item)));
    if (editingTierId === updatedTier.id) {
      applyFormValues(tierToFormValues(updatedTier));
    }
    toast.success(t("feedback.statusUpdated"));
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-6">
      {!embedded ? (
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="font-semibold text-2xl tracking-tight">{t("title")}</h1>
            <p className="mt-0.5 text-muted-foreground text-sm">{t("subtitle")}</p>
          </div>
          <Button size="sm" onClick={resetToNewTier}>
            {t("actions.newTier")}
          </Button>
        </div>
      ) : null}

      {loadFailed ? (
        <Card className="border-destructive/40">
          <CardContent className="pt-6 text-destructive text-sm">{t("states.loadFailed")}</CardContent>
        </Card>
      ) : null}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {[
          { label: t("summary.total"), value: tiers.length },
          { label: t("summary.active"), value: activeCount },
          { label: t("summary.inactive"), value: tiers.length - activeCount },
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
          {tiers.length === 0 ? (
            <div className="py-12 text-center text-muted-foreground text-sm">{t("states.empty")}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("table.name")}</TableHead>
                  <TableHead>{t("table.monthly")}</TableHead>
                  <TableHead>{t("table.annual")}</TableHead>
                  <TableHead>{t("table.currency")}</TableHead>
                  <TableHead>{t("table.features")}</TableHead>
                  <TableHead>{t("table.status")}</TableHead>
                  <TableHead>{t("table.updated")}</TableHead>
                  <TableHead>{t("table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tiers.map((tier) => (
                  <TableRow key={tier.id}>
                    <TableCell>
                      <div>
                        <p className="font-medium">{tier.name}</p>
                        <p className="max-w-sm text-muted-foreground text-xs">{tier.description || "—"}</p>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {formatMoney(tier.monthly_price, tier.currency)}
                    </TableCell>
                    <TableCell className="font-mono text-sm">{formatMoney(tier.annual_price, tier.currency)}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{tier.currency}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {tier.features.slice(0, 2).map((feature) => (
                          <Badge key={feature} variant="secondary" className="max-w-44 truncate">
                            {feature}
                          </Badge>
                        ))}
                        {tier.features.length > 2 ? (
                          <Badge variant="secondary">+{tier.features.length - 2}</Badge>
                        ) : null}
                        {tier.features.length === 0 ? <span className="text-muted-foreground text-xs">—</span> : null}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        className={
                          tier.is_active ? "bg-emerald-500/10 text-emerald-600" : "bg-muted text-muted-foreground"
                        }
                      >
                        {tier.is_active ? t("status.active") : t("status.inactive")}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground text-xs">
                      {hasHydrated ? formatDate(tier.updated_at) : "—"}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button variant="outline" size="sm" onClick={() => selectTier(tier)}>
                          {t("actions.edit")}
                        </Button>
                        <Button
                          variant={tier.is_active ? "destructive" : "default"}
                          size="sm"
                          disabled={pendingAction === `toggle:${tier.id}`}
                          onClick={() => void toggleTier(tier)}
                        >
                          {tier.is_active ? t("actions.deactivate") : t("actions.activate")}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{editingTierId ? t("form.titleEdit") : t("form.titleNew")}</CardTitle>
          <CardDescription>{t("form.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={saveTier}>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.name")}</p>
                <Controller
                  control={form.control}
                  name="name"
                  render={({ field }) => <Input {...field} value={field.value} />}
                />
                {fieldError(form.formState.errors.name)}
              </div>
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.currency")}</p>
                <Controller
                  control={form.control}
                  name="currency"
                  render={({ field }) => <Input {...field} className="w-24" maxLength={3} value={field.value} />}
                />
                {fieldError(form.formState.errors.currency)}
              </div>
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.monthlyPrice")}</p>
                <Controller
                  control={form.control}
                  name="monthly_price"
                  render={({ field }) => (
                    <Input {...field} inputMode="decimal" placeholder="9.99" value={field.value} />
                  )}
                />
                {fieldError(form.formState.errors.monthly_price)}
              </div>
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.annualPrice")}</p>
                <Controller
                  control={form.control}
                  name="annual_price"
                  render={({ field }) => (
                    <Input {...field} inputMode="decimal" placeholder="99.99" value={field.value} />
                  )}
                />
                {fieldError(form.formState.errors.annual_price)}
              </div>
              <div className="space-y-1 md:col-span-2">
                <p className="font-medium text-xs">{t("form.descriptionLabel")}</p>
                <Controller
                  control={form.control}
                  name="description"
                  render={({ field }) => <Textarea {...field} rows={3} value={field.value} />}
                />
              </div>
              <div className="space-y-1 md:col-span-2">
                <p className="font-medium text-xs">{t("form.features")}</p>
                <Controller
                  control={form.control}
                  name="features"
                  render={({ field }) => (
                    <Textarea {...field} rows={5} placeholder={t("form.featuresPlaceholder")} value={field.value} />
                  )}
                />
              </div>
            </div>

            <Controller
              control={form.control}
              name="is_active"
              render={({ field }) => (
                <div className="flex items-center gap-2">
                  <Switch checked={field.value} onCheckedChange={field.onChange} />
                  <span className="text-sm">{t("form.active")}</span>
                </div>
              )}
            />

            <div className="flex gap-2">
              <Button type="submit" size="sm" disabled={pendingAction === "save"}>
                {pendingAction === "save"
                  ? editingTierId
                    ? t("feedback.saving")
                    : t("feedback.creating")
                  : editingTierId
                    ? t("actions.save")
                    : t("actions.create")}
              </Button>
              <Button type="button" size="sm" variant="outline" onClick={resetToNewTier}>
                {embedded ? t("actions.newTier") : t("actions.cancel")}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
