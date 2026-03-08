"use client";

import { useEffect, useMemo, useState } from "react";

import { useTranslations } from "next-intl";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import {
  DEFAULT_PAYWALL_TEMPLATE,
  PAYWALL_TEMPLATES,
  type PaywallDefinition,
  parsePaywallDefinition,
  stringifyPaywallDefinition,
} from "@/lib/paywall-schema";
import { cn } from "@/lib/utils";

function normalizeHexColor(hex: string) {
  const clean = hex.replace("#", "");
  if (clean.length === 3) {
    return clean
      .split("")
      .map((part) => `${part}${part}`)
      .join("");
  }
  return clean;
}

function getContrastText(hex: string) {
  const normalized = normalizeHexColor(hex);
  const value = Number.parseInt(normalized, 16);
  const r = (value >> 16) & 255;
  const g = (value >> 8) & 255;
  const b = value & 255;
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return luminance > 0.62 ? "#0F172A" : "#FFFFFF";
}

function SummaryCard({ label, value }: { label: string; value: string | number }) {
  return (
    <Card>
      <CardContent className="pt-6">
        <p className="font-semibold text-muted-foreground text-xs uppercase tracking-widest">{label}</p>
        <p className="mt-2 font-bold text-2xl">{value}</p>
      </CardContent>
    </Card>
  );
}

function WebPreview({ paywall }: { paywall: PaywallDefinition }) {
  const mutedColor = paywall.theme.mode === "dark" ? "#94A3B8" : "#64748B";
  const borderColor = paywall.theme.mode === "dark" ? "rgba(148, 163, 184, 0.18)" : "rgba(15, 23, 42, 0.08)";
  const accentText = getContrastText(paywall.theme.accentColor);

  return (
    <div
      className="rounded-[28px] border p-6 shadow-sm"
      style={{
        backgroundColor: paywall.theme.backgroundColor,
        borderColor,
        color: paywall.theme.textColor,
      }}
    >
      <div className={cn("grid gap-6", paywall.layout === "split" ? "lg:grid-cols-[1.1fr_0.9fr]" : "grid-cols-1")}>
        <div className="space-y-5">
          {paywall.hero.badge ? (
            <Badge className="border-0" style={{ backgroundColor: paywall.theme.accentColor, color: accentText }}>
              {paywall.hero.badge}
            </Badge>
          ) : null}
          <div className="space-y-3">
            <h2 className="max-w-xl font-semibold text-4xl leading-tight">{paywall.hero.title}</h2>
            <p className="max-w-2xl text-base leading-7" style={{ color: mutedColor }}>
              {paywall.hero.subtitle}
            </p>
            {paywall.hero.socialProof ? <p className="font-medium text-sm">{paywall.hero.socialProof}</p> : null}
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            {paywall.features.map((feature) => (
              <div key={feature} className="rounded-2xl border px-4 py-3 text-sm" style={{ borderColor }}>
                <span className="font-medium">✓ {feature}</span>
              </div>
            ))}
          </div>
        </div>

        <div
          className={cn(
            "grid gap-4",
            paywall.plans.length === 3 ? "lg:grid-cols-1 xl:grid-cols-3" : "sm:grid-cols-2 lg:grid-cols-1",
          )}
        >
          {paywall.plans.map((plan) => (
            <div
              key={plan.id}
              className="rounded-3xl border p-5 shadow-sm"
              style={{
                backgroundColor: paywall.theme.surfaceColor,
                borderColor: plan.highlight ? paywall.theme.accentColor : borderColor,
                color: paywall.theme.textColor,
                boxShadow: plan.highlight ? `0 14px 40px ${paywall.theme.accentColor}22` : undefined,
              }}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-semibold text-lg">{plan.title}</p>
                  {plan.caption ? (
                    <p className="mt-1 text-sm" style={{ color: mutedColor }}>
                      {plan.caption}
                    </p>
                  ) : null}
                </div>
                {plan.badge ? (
                  <Badge className="border-0" style={{ backgroundColor: paywall.theme.accentColor, color: accentText }}>
                    {plan.badge}
                  </Badge>
                ) : null}
              </div>
              <div className="mt-6 flex items-end gap-1">
                <span className="font-bold text-4xl">{plan.price}</span>
                <span className="pb-1 text-sm" style={{ color: mutedColor }}>
                  {plan.period}
                </span>
              </div>
              <button
                className="mt-6 w-full rounded-xl px-4 py-3 font-semibold text-sm"
                style={{ backgroundColor: paywall.theme.accentColor, color: accentText }}
                type="button"
              >
                {paywall.cta.primaryLabel}
              </button>
            </div>
          ))}
        </div>
      </div>
      <Separator className="my-6 opacity-50" />
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <button className="text-left font-medium text-sm" style={{ color: mutedColor }} type="button">
          {paywall.footer.restoreLabel}
        </button>
        <div className="flex flex-col gap-2 text-sm sm:items-end">
          {paywall.cta.secondaryLabel ? (
            <button className="font-medium" style={{ color: mutedColor }} type="button">
              {paywall.cta.secondaryLabel}
            </button>
          ) : null}
          <p className="max-w-xl text-xs leading-5" style={{ color: mutedColor }}>
            {paywall.footer.legalText}
          </p>
        </div>
      </div>
    </div>
  );
}

function MobilePreview({ paywall }: { paywall: PaywallDefinition }) {
  const mutedColor = paywall.theme.mode === "dark" ? "#94A3B8" : "#64748B";
  const accentText = getContrastText(paywall.theme.accentColor);
  const borderColor = paywall.theme.mode === "dark" ? "rgba(148, 163, 184, 0.22)" : "rgba(15, 23, 42, 0.12)";

  return (
    <div className="mx-auto w-full max-w-[380px] rounded-[40px] border border-slate-700 bg-slate-950 p-3 shadow-2xl">
      <div
        className="rounded-[30px] p-5"
        style={{ backgroundColor: paywall.theme.backgroundColor, color: paywall.theme.textColor }}
      >
        <div className="mb-5 flex items-center justify-between">
          <div className="mx-auto h-1.5 w-24 rounded-full bg-white/15" />
          <span className="text-lg opacity-70">✕</span>
        </div>

        {paywall.hero.badge ? (
          <Badge className="mb-4 border-0" style={{ backgroundColor: paywall.theme.accentColor, color: accentText }}>
            {paywall.hero.badge}
          </Badge>
        ) : null}

        <div className="space-y-3">
          <h2 className="font-semibold text-3xl leading-tight">{paywall.hero.title}</h2>
          <p className="text-sm leading-6" style={{ color: mutedColor }}>
            {paywall.hero.subtitle}
          </p>
          {paywall.hero.socialProof ? (
            <p className="font-medium text-xs uppercase tracking-wide">{paywall.hero.socialProof}</p>
          ) : null}
        </div>

        <div className="mt-6 space-y-3">
          {paywall.features.map((feature) => (
            <div key={feature} className="flex items-center gap-3 rounded-2xl border px-4 py-3" style={{ borderColor }}>
              <span className="font-bold" style={{ color: paywall.theme.accentColor }}>
                ✓
              </span>
              <span className="text-sm">{feature}</span>
            </div>
          ))}
        </div>

        <div className="mt-6 space-y-3">
          {paywall.plans.map((plan) => (
            <div
              key={plan.id}
              className="rounded-3xl border p-4"
              style={{
                backgroundColor: paywall.theme.surfaceColor,
                borderColor: plan.highlight ? paywall.theme.accentColor : borderColor,
              }}
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="font-semibold text-base">{plan.title}</p>
                  {plan.caption ? (
                    <p className="mt-1 text-xs" style={{ color: mutedColor }}>
                      {plan.caption}
                    </p>
                  ) : null}
                </div>
                {plan.badge ? (
                  <Badge className="border-0" style={{ backgroundColor: paywall.theme.accentColor, color: accentText }}>
                    {plan.badge}
                  </Badge>
                ) : null}
              </div>
              <div className="mt-4 flex items-end gap-1">
                <span className="font-bold text-3xl">{plan.price}</span>
                <span className="pb-1 text-xs" style={{ color: mutedColor }}>
                  {plan.period}
                </span>
              </div>
              <button
                className="mt-4 w-full rounded-2xl px-4 py-3 font-semibold text-sm"
                style={{ backgroundColor: paywall.theme.accentColor, color: accentText }}
                type="button"
              >
                {paywall.cta.primaryLabel}
              </button>
            </div>
          ))}
        </div>

        <div className="mt-6 space-y-3 text-center">
          <button className="font-medium text-sm" style={{ color: mutedColor }} type="button">
            {paywall.footer.restoreLabel}
          </button>
          {paywall.cta.secondaryLabel ? (
            <button className="font-medium text-sm" style={{ color: mutedColor }} type="button">
              {paywall.cta.secondaryLabel}
            </button>
          ) : null}
          <p className="text-xs leading-5" style={{ color: mutedColor }}>
            {paywall.footer.legalText}
          </p>
        </div>
      </div>
    </div>
  );
}

export function PaywallCreatorPageClient() {
  const t = useTranslations("paywallCreator");
  const [schemaText, setSchemaText] = useState(() => stringifyPaywallDefinition(DEFAULT_PAYWALL_TEMPLATE));
  const parsed = useMemo(() => parsePaywallDefinition(schemaText), [schemaText]);
  const [previewConfig, setPreviewConfig] = useState(DEFAULT_PAYWALL_TEMPLATE);

  useEffect(() => {
    if (parsed.success) {
      setPreviewConfig(parsed.data);
    }
  }, [parsed]);

  function loadTemplate(template: keyof typeof PAYWALL_TEMPLATES) {
    setSchemaText(stringifyPaywallDefinition(PAYWALL_TEMPLATES[template]));
  }

  function formatJson() {
    if (parsed.success) {
      setSchemaText(stringifyPaywallDefinition(parsed.data));
      return;
    }
    try {
      setSchemaText(JSON.stringify(JSON.parse(schemaText), null, 2));
    } catch {
      // Ignore invalid JSON until the user fixes it.
    }
  }

  const schemaFields = ["id", "name", "platform", "layout", "theme", "hero", "features[]", "plans[]", "cta", "footer"];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 className="font-semibold text-2xl tracking-tight">{t("title")}</h1>
          <p className="mt-1 max-w-3xl text-muted-foreground text-sm">{t("subtitle")}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button size="sm" variant="outline" onClick={() => loadTemplate("mobileStreaming")}>
            {t("actions.mobileTemplate")}
          </Button>
          <Button size="sm" variant="outline" onClick={() => loadTemplate("webSaas")}>
            {t("actions.webTemplate")}
          </Button>
          <Button size="sm" onClick={formatJson}>
            {t("actions.format")}
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        <SummaryCard label={t("summary.validation")} value={parsed.success ? t("status.valid") : t("status.invalid")} />
        <SummaryCard label={t("summary.plans")} value={previewConfig.plans.length} />
        <SummaryCard label={t("summary.features")} value={previewConfig.features.length} />
      </div>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(420px,1.05fr)]">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t("editor.title")}</CardTitle>
            <CardDescription>{t("editor.description")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {!parsed.success ? (
              <Alert variant="destructive">
                <AlertTitle>{t("editor.validationFailed")}</AlertTitle>
                <AlertDescription>
                  <ul className="list-disc space-y-1 pl-5">
                    {parsed.errors.map((error) => (
                      <li key={error}>{error}</li>
                    ))}
                  </ul>
                </AlertDescription>
              </Alert>
            ) : (
              <Alert>
                <AlertTitle>{t("editor.validationSuccess")}</AlertTitle>
                <AlertDescription>{t("editor.validationSuccessDescription")}</AlertDescription>
              </Alert>
            )}

            <Textarea
              className="min-h-[640px] font-mono text-xs leading-6"
              spellCheck={false}
              value={schemaText}
              onChange={(event) => setSchemaText(event.target.value)}
            />
          </CardContent>
        </Card>

        <Tabs className="gap-4" defaultValue="web">
          <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h2 className="font-semibold text-lg">{t("preview.title")}</h2>
              <p className="text-muted-foreground text-sm">{t("preview.description")}</p>
            </div>
            <TabsList>
              <TabsTrigger value="web">{t("preview.web")}</TabsTrigger>
              <TabsTrigger value="mobile">{t("preview.mobile")}</TabsTrigger>
            </TabsList>
          </div>

          <Card>
            <CardContent className="space-y-4 pt-6">
              <div className="flex flex-wrap gap-2">
                <Badge variant="secondary">{previewConfig.platform}</Badge>
                <Badge variant="secondary">{previewConfig.layout}</Badge>
                <Badge variant="secondary">{previewConfig.theme.mode}</Badge>
                {!parsed.success ? <Badge variant="outline">{t("preview.fallback")}</Badge> : null}
              </div>

              <ScrollArea className="h-[760px] pr-4">
                <TabsContent value="web" className="mt-0">
                  <WebPreview paywall={previewConfig} />
                </TabsContent>
                <TabsContent value="mobile" className="mt-0">
                  <MobilePreview paywall={previewConfig} />
                </TabsContent>
              </ScrollArea>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-sm">{t("schema.title")}</CardTitle>
              <CardDescription>{t("schema.description")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-wrap gap-2">
                {schemaFields.map((field) => (
                  <Badge key={field} variant="outline">
                    {field}
                  </Badge>
                ))}
              </div>
              <Separator />
              <ul className="space-y-2 text-muted-foreground text-sm">
                <li>{t("schema.rules.colors")}</li>
                <li>{t("schema.rules.features")}</li>
                <li>{t("schema.rules.plans")}</li>
                <li>{t("schema.rules.highlight")}</li>
              </ul>
            </CardContent>
          </Card>
        </Tabs>
      </div>
    </div>
  );
}
