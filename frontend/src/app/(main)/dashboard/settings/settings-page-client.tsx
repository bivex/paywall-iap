"use client";

import { useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { changeAdminPasswordAction, updatePlatformSettings } from "@/actions/platform-settings";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import type { PlatformSettings } from "@/lib/platform-settings";

const formSchema = z
  .object({
    general: z.object({
      platform_name: z.string().trim().min(1),
      support_email: z.string().trim().email(),
      default_currency: z
        .string()
        .trim()
        .regex(/^[A-Za-z]{3}$/),
      dark_mode_default: z.boolean(),
    }),
    integrations: z.object({
      stripe_api_key: z.string(),
      stripe_webhook_secret: z.string(),
      stripe_test_mode: z.boolean(),
      apple_issuer_id: z.string(),
      apple_bundle_id: z.string(),
      google_service_account: z.string(),
      google_package_name: z.string(),
      matomo_url: z.union([z.literal(""), z.string().url()]),
      matomo_site_id: z.string(),
      matomo_auth_token: z.string(),
    }),
    notifications: z.object({
      new_subscription: z.boolean(),
      payment_failed: z.boolean(),
      subscription_cancelled: z.boolean(),
      refund_issued: z.boolean(),
      webhook_failed: z.boolean(),
      dunning_started: z.boolean(),
    }),
    security: z.object({
      jwt_expiry_hours: z.coerce.number().int().min(1).max(720),
      require_mfa: z.boolean(),
      enable_ip_allowlist: z.boolean(),
    }),
    currentPassword: z.string(),
    newPassword: z.string(),
    confirmPassword: z.string(),
  })
  .superRefine((value, ctx) => {
    const wantsPasswordChange = Boolean(value.currentPassword || value.newPassword || value.confirmPassword);
    if (!wantsPasswordChange) return;

    if (!value.currentPassword) {
      ctx.addIssue({ code: "custom", message: "Current password is required", path: ["currentPassword"] });
    }
    if (value.newPassword.length < 8) {
      ctx.addIssue({ code: "custom", message: "Password must be at least 8 characters", path: ["newPassword"] });
    }
    if (value.newPassword !== value.confirmPassword) {
      ctx.addIssue({ code: "custom", message: "Passwords do not match", path: ["confirmPassword"] });
    }
  });

type SettingsFormValues = z.infer<typeof formSchema>;
type SectionKey = "general" | "integrations" | "notifications" | "security" | "password";
type NotificationFieldName = keyof SettingsFormValues["notifications"];

const SECTION_FIELDS: Record<Exclude<SectionKey, "password">, Array<keyof SettingsFormValues | string>> = {
  general: ["general.platform_name", "general.support_email", "general.default_currency", "general.dark_mode_default"],
  integrations: [
    "integrations.stripe_api_key",
    "integrations.stripe_webhook_secret",
    "integrations.stripe_test_mode",
    "integrations.apple_issuer_id",
    "integrations.apple_bundle_id",
    "integrations.google_service_account",
    "integrations.google_package_name",
    "integrations.matomo_url",
    "integrations.matomo_site_id",
    "integrations.matomo_auth_token",
  ],
  notifications: [
    "notifications.new_subscription",
    "notifications.payment_failed",
    "notifications.subscription_cancelled",
    "notifications.refund_issued",
    "notifications.webhook_failed",
    "notifications.dunning_started",
  ],
  security: ["security.jwt_expiry_hours", "security.require_mfa", "security.enable_ip_allowlist"],
};

function toPayload(values: SettingsFormValues): PlatformSettings {
  return {
    general: {
      ...values.general,
      default_currency: values.general.default_currency.toUpperCase(),
    },
    integrations: values.integrations,
    notifications: values.notifications,
    security: {
      ...values.security,
      jwt_expiry_hours: Number(values.security.jwt_expiry_hours),
    },
  };
}

function fieldError(error?: { message?: string }) {
  if (!error?.message) return null;
  return <p className="text-destructive text-xs">{error.message}</p>;
}

export function SettingsPageClient({ initialSettings }: { initialSettings: PlatformSettings }) {
  const t = useTranslations("settings");
  const [pendingSection, setPendingSection] = useState<SectionKey | null>(null);
  const notificationItems: Array<{ name: NotificationFieldName; label: string }> = [
    { name: "new_subscription", label: t("notifications.newSubscription") },
    { name: "payment_failed", label: t("notifications.paymentFailed") },
    { name: "subscription_cancelled", label: t("notifications.subscriptionCancelled") },
    { name: "refund_issued", label: t("notifications.refundIssued") },
    { name: "webhook_failed", label: t("notifications.webhookFailed") },
    { name: "dunning_started", label: t("notifications.dunningStarted") },
  ];

  const form = useForm<SettingsFormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      ...initialSettings,
      currentPassword: "",
      newPassword: "",
      confirmPassword: "",
    },
  });

  const saveSection = async (section: Exclude<SectionKey, "password">) => {
    const isValid = await form.trigger(SECTION_FIELDS[section] as never[], { shouldFocus: true });
    if (!isValid) return;

    setPendingSection(section);
    const result = await updatePlatformSettings(toPayload(form.getValues()));
    setPendingSection(null);

    if (!result.ok || !result.data) {
      toast.error(result.error ?? t("feedback.saveFailed"));
      return;
    }

    form.reset({ ...result.data, currentPassword: "", newPassword: "", confirmPassword: "" });
    toast.success(t("feedback.settingsSaved"));
  };

  const updatePassword = async () => {
    const isValid = await form.trigger(["currentPassword", "newPassword", "confirmPassword"], { shouldFocus: true });
    if (!isValid) return;

    setPendingSection("password");
    const values = form.getValues();
    const result = await changeAdminPasswordAction({
      currentPassword: values.currentPassword,
      newPassword: values.newPassword,
      confirmPassword: values.confirmPassword,
    });
    setPendingSection(null);

    if (!result.ok) {
      toast.error(result.error ?? t("feedback.passwordFailed"));
      return;
    }

    form.setValue("currentPassword", "");
    form.setValue("newPassword", "");
    form.setValue("confirmPassword", "");
    toast.success(t("feedback.passwordSaved"));
  };

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-semibold text-2xl">{t("title")}</h1>
      <Tabs defaultValue="general">
        <TabsList>
          <TabsTrigger value="general">{t("tabs.general")}</TabsTrigger>
          <TabsTrigger value="integrations">{t("tabs.integrations")}</TabsTrigger>
          <TabsTrigger value="notifications">{t("tabs.notifications")}</TabsTrigger>
          <TabsTrigger value="security">{t("tabs.security")}</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("general.title")}</CardTitle>
              <CardDescription>{t("descriptions.general")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("general.platformName")}</p>
                <Input {...form.register("general.platform_name")} />
                {fieldError(form.formState.errors.general?.platform_name)}
              </div>
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("general.supportEmail")}</p>
                <Input type="email" {...form.register("general.support_email")} />
                {fieldError(form.formState.errors.general?.support_email)}
              </div>
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("general.defaultCurrency")}</p>
                <Input className="w-24" maxLength={3} {...form.register("general.default_currency")} />
                {fieldError(form.formState.errors.general?.default_currency)}
              </div>
              <Controller
                control={form.control}
                name="general.dark_mode_default"
                render={({ field }) => (
                  <div className="flex items-center gap-2">
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                    <span className="text-sm">{t("general.darkMode")}</span>
                  </div>
                )}
              />
              <Button size="sm" onClick={() => void saveSection("general")} disabled={pendingSection === "general"}>
                {pendingSection === "general" ? t("feedback.saving") : t("general.save")}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="integrations" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("tabs.integrations")}</CardTitle>
              <CardDescription>{t("descriptions.integrations")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("stripe.apiKey")}</p>
                  <Input type="password" {...form.register("integrations.stripe_api_key")} />
                </div>
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("stripe.webhookSecret")}</p>
                  <Input type="password" {...form.register("integrations.stripe_webhook_secret")} />
                </div>
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("apple.issuerId")}</p>
                  <Input {...form.register("integrations.apple_issuer_id")} />
                </div>
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("apple.bundleId")}</p>
                  <Input {...form.register("integrations.apple_bundle_id")} />
                </div>
                <div className="space-y-1 md:col-span-2">
                  <p className="font-medium text-xs">{t("google.serviceAccount")}</p>
                  <Textarea rows={5} {...form.register("integrations.google_service_account")} />
                </div>
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("google.packageName")}</p>
                  <Input {...form.register("integrations.google_package_name")} />
                </div>
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("matomo.url")}</p>
                  <Input {...form.register("integrations.matomo_url")} />
                  {fieldError(form.formState.errors.integrations?.matomo_url)}
                </div>
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("matomo.siteId")}</p>
                  <Input {...form.register("integrations.matomo_site_id")} />
                </div>
                <div className="space-y-1">
                  <p className="font-medium text-xs">{t("matomo.authToken")}</p>
                  <Input type="password" {...form.register("integrations.matomo_auth_token")} />
                </div>
              </div>
              <Controller
                control={form.control}
                name="integrations.stripe_test_mode"
                render={({ field }) => (
                  <div className="flex items-center gap-2">
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                    <span className="text-sm">{t("stripe.testMode")}</span>
                  </div>
                )}
              />
              <Button
                size="sm"
                onClick={() => void saveSection("integrations")}
                disabled={pendingSection === "integrations"}
              >
                {pendingSection === "integrations" ? t("feedback.saving") : t("actions.saveIntegrations")}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="notifications" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("notifications.title")}</CardTitle>
              <CardDescription>{t("descriptions.notifications")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              {notificationItems.map(({ name, label }) => (
                <Controller
                  key={name}
                  control={form.control}
                  name={`notifications.${name}` as const}
                  render={({ field }) => (
                    <div className="flex items-center gap-2">
                      <Switch checked={Boolean(field.value)} onCheckedChange={field.onChange} />
                      <span className="text-sm">{label}</span>
                    </div>
                  )}
                />
              ))}
              <Button
                size="sm"
                className="mt-2"
                onClick={() => void saveSection("notifications")}
                disabled={pendingSection === "notifications"}
              >
                {pendingSection === "notifications" ? t("feedback.saving") : t("notifications.save")}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("security.title")}</CardTitle>
              <CardDescription>{t("descriptions.security")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="space-y-1">
                <p className="font-medium text-xs">{t("security.jwtExpiry")}</p>
                <Input
                  type="number"
                  className="w-28"
                  {...form.register("security.jwt_expiry_hours", { valueAsNumber: true })}
                />
                {fieldError(form.formState.errors.security?.jwt_expiry_hours as { message?: string } | undefined)}
              </div>
              <Controller
                control={form.control}
                name="security.require_mfa"
                render={({ field }) => (
                  <div className="flex items-center gap-2">
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                    <span className="text-sm">{t("security.mfa")}</span>
                  </div>
                )}
              />
              <Controller
                control={form.control}
                name="security.enable_ip_allowlist"
                render={({ field }) => (
                  <div className="flex items-center gap-2">
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                    <span className="text-sm">{t("security.ipAllowlist")}</span>
                  </div>
                )}
              />
              <Button size="sm" onClick={() => void saveSection("security")} disabled={pendingSection === "security"}>
                {pendingSection === "security" ? t("feedback.saving") : t("actions.saveSecurity")}
              </Button>
              <Separator />
              <p className="font-medium text-xs">{t("security.changePassword")}</p>
              <Input
                placeholder={t("security.currentPassword")}
                type="password"
                {...form.register("currentPassword")}
              />
              {fieldError(form.formState.errors.currentPassword)}
              <Input placeholder={t("security.newPassword")} type="password" {...form.register("newPassword")} />
              {fieldError(form.formState.errors.newPassword)}
              <Input
                placeholder={t("security.confirmPassword")}
                type="password"
                {...form.register("confirmPassword")}
              />
              {fieldError(form.formState.errors.confirmPassword)}
              <Button
                size="sm"
                variant="destructive"
                onClick={() => void updatePassword()}
                disabled={pendingSection === "password"}
              >
                {pendingSection === "password" ? t("feedback.saving") : t("security.updatePassword")}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
