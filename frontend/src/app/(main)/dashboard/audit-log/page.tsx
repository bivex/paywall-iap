import { getTranslations } from "next-intl/server";
import { Suspense } from "react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { getAuditLog } from "@/actions/audit-log";
import { AuditLogFilters } from "./_components/audit-log-filters";

// Badge style per action type
const actionMeta: Record<string, { label: string; className: string }> = {
  grant_subscription:   { label: "Grant",      className: "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400" },
  revoke_subscription:  { label: "Revoke",     className: "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400" },
  grant_refund:         { label: "Refund",     className: "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400" },
  cancel_subscription:  { label: "Cancel",     className: "bg-rose-100 text-rose-800 dark:bg-rose-900/30 dark:text-rose-400" },
  update_plan_price:    { label: "Pricing",    className: "bg-violet-100 text-violet-800 dark:bg-violet-900/30 dark:text-violet-400" },
  trigger_dunning:      { label: "Dunning",    className: "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400" },
  replay_webhook:       { label: "Webhook",    className: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400" },
  manual_renewal:       { label: "Renewal",    className: "bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-400" },
};

function ActionBadge({ action }: { action: string }) {
  const meta = actionMeta[action];
  return (
    <Badge className={meta?.className ?? "bg-gray-100 text-gray-700"}>
      {meta?.label ?? action}
    </Badge>
  );
}

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString("en-US", {
      month: "short", day: "numeric", year: "numeric",
      hour: "2-digit", minute: "2-digit", hour12: false,
    });
  } catch {
    return iso;
  }
}

interface SearchParams {
  page?: string;
  action?: string;
  search?: string;
  from?: string;
  to?: string;
}

export default async function AuditLogPage({
  searchParams,
}: {
  searchParams: Promise<SearchParams>;
}) {
  const t = await getTranslations("auditLog");
  const sp = await searchParams;

  const page = Math.max(1, parseInt(sp.page ?? "1", 10));
  const data = await getAuditLog({
    page,
    limit: 20,
    action: sp.action,
    search: sp.search,
    from: sp.from,
    to: sp.to,
  });

  const buildPageUrl = (p: number) => {
    const params = new URLSearchParams();
    if (sp.action) params.set("action", sp.action);
    if (sp.search) params.set("search", sp.search);
    if (sp.from) params.set("from", sp.from);
    if (sp.to) params.set("to", sp.to);
    params.set("page", String(p));
    return `/dashboard/audit-log?${params.toString()}`;
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between flex-wrap gap-2">
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <Button variant="outline" size="sm" asChild>
          <Link href={`/api/proxy/audit-log/export?action=${sp.action ?? ""}&search=${sp.search ?? ""}&from=${sp.from ?? ""}&to=${sp.to ?? ""}`}>
            {t("exportCsv")}
          </Link>
        </Button>
      </div>

      <Card>
        <CardContent className="pt-4 space-y-4">
          {/* Client-side filter controls (updates URL) */}
          <Suspense fallback={null}>
            <AuditLogFilters />
          </Suspense>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-40">{t("table.timestamp")}</TableHead>
                <TableHead>{t("table.admin")}</TableHead>
                <TableHead className="w-36">{t("table.action")}</TableHead>
                <TableHead>{t("table.target")}</TableHead>
                <TableHead className="w-36">{t("table.ipAddress")}</TableHead>
                <TableHead>{t("table.detail")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-10">
                    No audit log entries found.
                  </TableCell>
                </TableRow>
              ) : (
                data.rows.map((row) => (
                  <TableRow key={row.ID}>
                    <TableCell className="text-xs text-muted-foreground whitespace-nowrap">
                      {formatDate(row.Time)}
                    </TableCell>
                    <TableCell className="text-sm font-medium">{row.AdminEmail}</TableCell>
                    <TableCell><ActionBadge action={row.Action} /></TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {row.TargetType}
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {row.IPAddress || "—"}
                    </TableCell>
                    <TableCell className="text-xs max-w-xs truncate" title={row.Detail}>
                      {row.Detail || "—"}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {/* Pagination */}
          <div className="flex items-center justify-between pt-2">
            <span className="text-xs text-muted-foreground">
              {data.total} {data.total === 1 ? "entry" : "entries"}
              {data.total_pages > 1 && ` · Page ${data.page} of ${data.total_pages}`}
            </span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                asChild={page > 1}
              >
                {page > 1 ? (
                  <Link href={buildPageUrl(page - 1)}>← Previous</Link>
                ) : (
                  <span>← Previous</span>
                )}
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= data.total_pages}
                asChild={page < data.total_pages}
              >
                {page < data.total_pages ? (
                  <Link href={buildPageUrl(page + 1)}>Next →</Link>
                ) : (
                  <span>Next →</span>
                )}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

