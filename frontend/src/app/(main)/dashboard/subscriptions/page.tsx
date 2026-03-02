/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-02 06:33
 * Last Updated: 2026-03-02 06:33
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

import { Suspense } from "react";
import { getTranslations } from "next-intl/server";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { getSubscriptions } from "@/actions/subscriptions";
import type { SubscriptionsParams } from "@/actions/subscriptions";
import { formatSource, formatPlanType } from "@/lib/subscriptions/format";
import { SubscriptionsFilters } from "./_components/subscriptions-filters";

const statusClassMap: Record<string, string> = {
  active: "bg-green-100 text-green-800",
  grace: "bg-yellow-100 text-yellow-800",
  cancelled: "bg-orange-100 text-orange-800",
  expired: "bg-red-100 text-red-800",
};

const PAGE_SIZE = 20;

interface Props {
  searchParams: Promise<Record<string, string | undefined>>;
}

export default async function SubscriptionsPage({ searchParams }: Props) {
  const [t, sp] = await Promise.all([
    getTranslations("subscriptions"),
    searchParams,
  ]);

  const page = Math.max(1, parseInt(sp.page ?? "1", 10) || 1);

  const params: SubscriptionsParams = {
    page,
    limit: PAGE_SIZE,
    status: sp.status,
    source: sp.source,
    platform: sp.platform,
    plan_type: sp.plan_type,
    search: sp.search,
    date_from: sp.date_from,
    date_to: sp.date_to,
  };

  const data = await getSubscriptions(params);
  const notAuthed = data === null;
  const subs = data?.subscriptions ?? [];
  const total = data?.total ?? 0;
  const totalPages = data?.total_pages ?? 1;

  const buildPageUrl = (p: number) => {
    const qs = new URLSearchParams();
    if (sp.status) qs.set("status", sp.status);
    if (sp.source) qs.set("source", sp.source);
    if (sp.platform) qs.set("platform", sp.platform);
    if (sp.plan_type) qs.set("plan_type", sp.plan_type);
    if (sp.search) qs.set("search", sp.search);
    if (sp.date_from) qs.set("date_from", sp.date_from);
    if (sp.date_to) qs.set("date_to", sp.date_to);
    qs.set("page", String(p));
    return `?${qs.toString()}`;
  };

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <Card>
        <CardContent className="pt-4 space-y-4">
          <Suspense>
            <SubscriptionsFilters />
          </Suspense>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("table.user")}</TableHead>
                <TableHead>{t("table.status")}</TableHead>
                <TableHead>{t("table.source")}</TableHead>
                <TableHead>{t("table.plan")}</TableHead>
                <TableHead>{t("table.expires")}</TableHead>
                <TableHead>{t("table.ltv")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {subs.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    {notAuthed ? "⚠️ Not authenticated — please log in." : "No subscriptions found."}
                  </TableCell>
                </TableRow>
              ) : (
                subs.map((s) => (
                  <TableRow key={s.id}>
                    <TableCell className="font-medium">{s.email || s.user_id}</TableCell>
                    <TableCell>
                      <Badge className={statusClassMap[s.status] ?? "bg-muted text-muted-foreground"}>
                        {s.status}
                      </Badge>
                    </TableCell>
                    <TableCell>{formatSource(s.source, s.platform)}</TableCell>
                    <TableCell>{formatPlanType(s.plan_type)}</TableCell>
                    <TableCell>{new Date(s.expires_at).toLocaleDateString("en-US", { year: "numeric", month: "short", day: "numeric" })}</TableCell>
                    <TableCell>${s.ltv.toFixed(2)}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {/* Pagination */}
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {total > 0
                ? `Showing ${(page - 1) * PAGE_SIZE + 1}–${Math.min(page * PAGE_SIZE, total)} of ${total}`
                : "No results"}
            </span>
            <div className="flex gap-1">
              {page > 1 && (
                <Button variant="outline" size="sm" asChild>
                  <a href={buildPageUrl(page - 1)}>← Prev</a>
                </Button>
              )}
              {page < totalPages && (
                <Button variant="outline" size="sm" asChild>
                  <a href={buildPageUrl(page + 1)}>Next →</a>
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
