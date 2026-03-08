"use client";

import { useState } from "react";

import { useRouter } from "next/navigation";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { deactivateWinbackCampaignAction, launchWinbackCampaignAction } from "@/actions/winback";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  EMPTY_WINBACK_CAMPAIGN_INPUT,
  type LaunchWinbackCampaignInput,
  type WinbackCampaign,
  type WinbackDiscountType,
} from "@/lib/winback";

const formSchema = z
  .object({
    campaign_id: z.string().trim().min(1),
    discount_type: z.enum(["percentage", "fixed"]),
    discount_value: z.string().trim(),
    duration_days: z.string().trim(),
    days_since_churn: z.string().trim(),
  })
  .superRefine((value, ctx) => {
    const numericFields = [
      ["discount_value", value.discount_value],
      ["duration_days", value.duration_days],
      ["days_since_churn", value.days_since_churn],
    ] as const;

    for (const [field, raw] of numericFields) {
      const parsed = Number(raw);
      if (!Number.isFinite(parsed) || parsed <= 0) {
        ctx.addIssue({ code: "custom", message: "Must be greater than zero", path: [field] });
      }
    }

    if (value.discount_type === "percentage") {
      const parsed = Number(value.discount_value);
      if (Number.isFinite(parsed) && parsed > 100) {
        ctx.addIssue({ code: "custom", message: "Percentage cannot exceed 100", path: ["discount_value"] });
      }
    }
  });

type WinbackFormValues = z.infer<typeof formSchema>;

const EMPTY_FORM_VALUES: WinbackFormValues = {
  campaign_id: EMPTY_WINBACK_CAMPAIGN_INPUT.campaign_id,
  discount_type: EMPTY_WINBACK_CAMPAIGN_INPUT.discount_type,
  discount_value: EMPTY_WINBACK_CAMPAIGN_INPUT.discount_value.toString(),
  duration_days: EMPTY_WINBACK_CAMPAIGN_INPUT.duration_days.toString(),
  days_since_churn: EMPTY_WINBACK_CAMPAIGN_INPUT.days_since_churn.toString(),
};

function fieldError(error?: { message?: string }) {
  if (!error?.message) return null;
  return <p className="text-destructive text-xs">{error.message}</p>;
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatDiscount(campaign: WinbackCampaign, t: ReturnType<typeof useTranslations>) {
  if (campaign.discount_type === "percentage") {
    return `${campaign.discount_value}%`;
  }
  return `${t("discount.fixedShort")} ${campaign.discount_value.toFixed(2)}`;
}

function toPayload(values: WinbackFormValues): LaunchWinbackCampaignInput {
  return {
    campaign_id: values.campaign_id.trim(),
    discount_type: values.discount_type,
    discount_value: Number(values.discount_value),
    duration_days: Number(values.duration_days),
    days_since_churn: Number(values.days_since_churn),
  };
}

export function WinbackPageClient({
  initialCampaigns,
  loadFailed,
}: {
  initialCampaigns: WinbackCampaign[];
  loadFailed: boolean;
}) {
  const t = useTranslations("winback");
  const router = useRouter();
  const [campaigns, setCampaigns] = useState(initialCampaigns);
  const [pendingLaunch, setPendingLaunch] = useState(false);
  const [pendingCampaignAction, setPendingCampaignAction] = useState<string | null>(null);
  const form = useForm<WinbackFormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: EMPTY_FORM_VALUES,
  });

  const totalOffers = campaigns.reduce((sum, campaign) => sum + campaign.total_offers, 0);
  const activeOffers = campaigns.reduce((sum, campaign) => sum + campaign.active_offers, 0);
  const acceptedOffers = campaigns.reduce((sum, campaign) => sum + campaign.accepted_offers, 0);

  const launchCampaign = form.handleSubmit(async (values) => {
    setPendingLaunch(true);
    const result = await launchWinbackCampaignAction(toPayload(values));
    setPendingLaunch(false);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.launchFailed"));
      return;
    }

    const launchedCampaign = result.data;
    setCampaigns((current) => [
      launchedCampaign,
      ...current.filter((item) => item.campaign_id !== launchedCampaign.campaign_id),
    ]);
    form.reset(EMPTY_FORM_VALUES);
    toast.success(t("feedback.campaignLaunched"));
    router.refresh();
  });

  async function deactivateCampaign(campaignId: string) {
    setPendingCampaignAction(campaignId);
    const result = await deactivateWinbackCampaignAction(campaignId);
    setPendingCampaignAction(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.deactivateFailed"));
      return;
    }

    const updatedCampaign = result.data;
    setCampaigns((current) =>
      current.map((campaign) => (campaign.campaign_id === updatedCampaign.campaign_id ? updatedCampaign : campaign)),
    );
    toast.success(t("feedback.campaignDeactivated"));
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

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {[
          { label: t("summary.campaigns"), value: campaigns.length },
          { label: t("summary.offers"), value: totalOffers },
          { label: t("summary.activeOffers"), value: activeOffers },
          { label: t("summary.acceptedOffers"), value: acceptedOffers },
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
          {campaigns.length === 0 ? (
            <div className="py-12 text-center text-muted-foreground text-sm">{t("states.empty")}</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("table.campaignId")}</TableHead>
                  <TableHead>{t("table.discount")}</TableHead>
                  <TableHead>{t("table.total")}</TableHead>
                  <TableHead>{t("table.active")}</TableHead>
                  <TableHead>{t("table.accepted")}</TableHead>
                  <TableHead>{t("table.expired")}</TableHead>
                  <TableHead>{t("table.declined")}</TableHead>
                  <TableHead>{t("table.launched")}</TableHead>
                  <TableHead>{t("table.expires")}</TableHead>
                  <TableHead>{t("table.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {campaigns.map((campaign) => (
                  <TableRow key={campaign.campaign_id}>
                    <TableCell>
                      <Badge variant="outline">{campaign.campaign_id}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-1">
                        <span className="font-medium">{formatDiscount(campaign, t)}</span>
                        <span className="text-muted-foreground text-xs">
                          {campaign.discount_type === "percentage" ? t("discount.percentage") : t("discount.fixed")}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-sm">{campaign.total_offers}</TableCell>
                    <TableCell className="font-mono text-sm">{campaign.active_offers}</TableCell>
                    <TableCell className="font-mono text-sm">{campaign.accepted_offers}</TableCell>
                    <TableCell className="font-mono text-sm">{campaign.expired_offers}</TableCell>
                    <TableCell className="font-mono text-sm">{campaign.declined_offers}</TableCell>
                    <TableCell className="text-muted-foreground text-xs">{formatDate(campaign.launched_at)}</TableCell>
                    <TableCell className="text-muted-foreground text-xs">
                      {formatDate(campaign.latest_expiry_at)}
                    </TableCell>
                    <TableCell>
                      {campaign.active_offers > 0 ? (
                        <Button
                          type="button"
                          size="sm"
                          variant="outline"
                          disabled={pendingCampaignAction !== null}
                          onClick={() => void deactivateCampaign(campaign.campaign_id)}
                        >
                          {pendingCampaignAction === campaign.campaign_id
                            ? t("feedback.deactivating")
                            : t("actions.deactivate")}
                        </Button>
                      ) : (
                        <span className="text-muted-foreground text-xs">—</span>
                      )}
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
          <CardTitle className="text-sm">{t("form.title")}</CardTitle>
          <CardDescription>{t("form.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={launchCampaign}>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.campaignId")}</p>
                <Input placeholder={t("form.campaignIdPlaceholder")} {...form.register("campaign_id")} />
                {fieldError(form.formState.errors.campaign_id)}
              </div>

              <Controller
                control={form.control}
                name="discount_type"
                render={({ field }) => (
                  <div className="space-y-1">
                    <p className="font-medium text-xs">{t("form.discountType")}</p>
                    <Select value={field.value} onValueChange={(value) => field.onChange(value as WinbackDiscountType)}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="percentage">{t("discount.percentage")}</SelectItem>
                        <SelectItem value="fixed">{t("discount.fixed")}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                )}
              />

              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.discountValue")}</p>
                <Input inputMode="decimal" placeholder="25" {...form.register("discount_value")} />
                {fieldError(form.formState.errors.discount_value)}
              </div>

              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.durationDays")}</p>
                <Input inputMode="numeric" placeholder="14" {...form.register("duration_days")} />
                {fieldError(form.formState.errors.duration_days)}
              </div>

              <div className="space-y-1">
                <p className="font-medium text-xs">{t("form.daysSinceChurn")}</p>
                <Input inputMode="numeric" placeholder="30" {...form.register("days_since_churn")} />
                {fieldError(form.formState.errors.days_since_churn)}
              </div>
            </div>

            <div className="flex gap-2">
              <Button type="submit" size="sm" disabled={pendingLaunch}>
                {pendingLaunch ? t("feedback.launching") : t("actions.launch")}
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
