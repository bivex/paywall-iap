"use client";

import { useMemo, useState } from "react";

import Link from "next/link";

import { ExternalLink, Settings2 } from "lucide-react";
import { useTranslations } from "next-intl";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

type DateRangeKey = "last7" | "last30" | "last90";

const DATE_QUERY: Record<DateRangeKey, string> = {
  last7: "last7",
  last30: "last30",
  last90: "last90",
};

function normalizeBaseUrl(url: string) {
  return url.replace(/\/$/, "");
}

function buildMatomoDashboardUrl(baseUrl: string, siteId: string, range: DateRangeKey) {
  const url = new URL(`${normalizeBaseUrl(baseUrl)}/index.php`);
  url.searchParams.set("module", "CoreHome");
  url.searchParams.set("action", "index");
  url.searchParams.set("idSite", siteId);
  url.searchParams.set("period", "range");
  url.searchParams.set("date", DATE_QUERY[range]);
  return url.toString();
}

function maskUrl(value: string) {
  try {
    const url = new URL(value);
    return `${url.origin}${url.pathname}`;
  } catch {
    return value;
  }
}

export function MatomoPageClient({
  config,
}: {
  config: {
    url: string;
    siteId: string;
    hasAuthToken: boolean;
  };
}) {
  const t = useTranslations("matomo");
  const [range, setRange] = useState<DateRangeKey>("last30");

  const isConfigured = Boolean(config.url && config.siteId);
  const dashboardUrl = useMemo(() => {
    if (!isConfigured) return "";
    return buildMatomoDashboardUrl(config.url, config.siteId, range);
  }, [config.siteId, config.url, isConfigured, range]);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-semibold text-2xl tracking-tight">{t("title")}</h1>
          <p className="mt-0.5 text-muted-foreground text-sm">{t("subtitle")}</p>
        </div>

        <div className="flex gap-2">
          <Select value={range} onValueChange={(value) => setRange(value as DateRangeKey)}>
            <SelectTrigger className="w-44">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="last7">{t("ranges.last7")}</SelectItem>
              <SelectItem value="last30">{t("ranges.last30")}</SelectItem>
              <SelectItem value="last90">{t("ranges.last90")}</SelectItem>
            </SelectContent>
          </Select>

          <Button asChild variant="outline" size="sm">
            <Link prefetch={false} href="/dashboard/settings">
              <Settings2 className="size-4" />
              {t("actions.openSettings")}
            </Link>
          </Button>

          <Button asChild size="sm" disabled={!isConfigured}>
            <a href={dashboardUrl || undefined} target="_blank" rel="noreferrer">
              <ExternalLink className="size-4" />
              {t("actions.openMatomo")}
            </a>
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {[
          { label: t("summary.status"), value: isConfigured ? t("summary.configured") : t("summary.notConfigured") },
          { label: t("summary.siteId"), value: config.siteId || "—" },
          { label: t("summary.authToken"), value: config.hasAuthToken ? t("summary.saved") : t("summary.missing") },
        ].map((item) => (
          <Card key={item.label}>
            <CardContent className="pt-6">
              <p className="font-semibold text-muted-foreground text-xs uppercase tracking-widest">{item.label}</p>
              <p className="mt-2 break-all font-bold text-lg">{item.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("connection.title")}</CardTitle>
          <CardDescription>{t("connection.description")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-2">
          <div className="rounded-md border p-4">
            <p className="font-medium text-sm">{t("connection.urlLabel")}</p>
            <p className="mt-1 text-muted-foreground text-sm">
              {config.url ? maskUrl(config.url) : t("connection.notSet")}
            </p>
          </div>
          <div className="rounded-md border p-4">
            <p className="font-medium text-sm">{t("connection.behaviorTitle")}</p>
            <p className="mt-1 text-muted-foreground text-sm">{t("connection.behaviorBody")}</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{t("embed.title")}</CardTitle>
          <CardDescription>{t("embed.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          {!isConfigured ? (
            <div className="flex min-h-80 items-center justify-center rounded-md border-2 border-muted-foreground/30 border-dashed bg-muted/20 p-6 text-center">
              <div className="space-y-2">
                <p className="font-medium text-sm">{t("embed.notConfiguredTitle")}</p>
                <p className="text-muted-foreground text-sm">{t("embed.notConfiguredBody")}</p>
                <Button asChild variant="outline" size="sm">
                  <Link prefetch={false} href="/dashboard/settings">
                    {t("embed.configure")}
                  </Link>
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-3">
              <div className="rounded-md border bg-muted/20 px-4 py-3 text-muted-foreground text-sm">
                {t("embed.bestEffortNotice")}
              </div>
              <iframe
                key={dashboardUrl}
                title="Matomo dashboard"
                src={dashboardUrl}
                className="h-[720px] w-full rounded-md border bg-background"
              />
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
