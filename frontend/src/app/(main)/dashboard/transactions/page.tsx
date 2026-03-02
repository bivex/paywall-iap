import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { AlertCircle, CheckCircle2, XCircle, RefreshCw, DollarSign, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";
import { getTransactions } from "@/actions/transactions";
import type { TransactionsParams } from "@/actions/transactions";
import { formatSource } from "@/lib/subscriptions/format";
import { TransactionsFilters } from "./_components/transactions-filters";
import { Suspense } from "react";

const PAGE_SIZE = 20;

const STATUS_STYLE: Record<string, string> = {
  success:  "bg-emerald-500/10 text-emerald-600 border-emerald-500/20",
  failed:   "bg-red-500/10 text-red-600 border-red-500/20",
  refunded: "bg-slate-500/10 text-slate-600 border-slate-500/20",
};

function fmt(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    month: "short", day: "numeric", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

interface Props {
  searchParams: Promise<Record<string, string | undefined>>;
}

export default async function TransactionsPage({ searchParams }: Props) {
  const sp = await searchParams;
  const page = Math.max(1, parseInt(sp.page ?? "1", 10) || 1);

  const params: TransactionsParams = {
    page,
    limit: PAGE_SIZE,
    status: sp.status,
    source: sp.source,
    platform: sp.platform,
    search: sp.search,
    date_from: sp.date_from,
    date_to: sp.date_to,
  };

  const data = await getTransactions(params);

  if (!data) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-24 text-muted-foreground">
        <AlertCircle className="h-8 w-8" />
        <p className="text-sm">Failed to load — make sure you are logged in.</p>
      </div>
    );
  }

  const { transactions, summary, total, total_pages: totalPages } = data;

  const buildPageUrl = (p: number) => {
    const qs = new URLSearchParams();
    if (sp.status) qs.set("status", sp.status);
    if (sp.source) qs.set("source", sp.source);
    if (sp.platform) qs.set("platform", sp.platform);
    if (sp.search) qs.set("search", sp.search);
    if (sp.date_from) qs.set("date_from", sp.date_from);
    if (sp.date_to) qs.set("date_to", sp.date_to);
    qs.set("page", String(p));
    return `?${qs.toString()}`;
  };

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Transaction Reconciliation</h1>
        <p className="text-sm text-muted-foreground mt-0.5">
          Full transaction ledger · reconcile payments across IAP, Stripe & Google Play
        </p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
        {[
          { label: "Total",      value: summary.total_count,    icon: DollarSign,   color: "text-foreground",      bg: "bg-muted"              },
          { label: "Success",    value: summary.success_count,  icon: CheckCircle2, color: "text-emerald-500",     bg: "bg-emerald-500/10"     },
          { label: "Failed",     value: summary.failed_count,   icon: XCircle,      color: "text-red-500",         bg: "bg-red-500/10"         },
          { label: "Refunded",   value: summary.refunded_count, icon: RefreshCw,    color: "text-slate-500",       bg: "bg-slate-500/10"       },
          { label: "Revenue",    value: `$${summary.total_revenue.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`, icon: DollarSign, color: "text-emerald-500", bg: "bg-emerald-500/10" },
          { label: "Refunds",    value: `$${summary.total_refunded.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`, icon: RefreshCw,  color: "text-slate-500",  bg: "bg-slate-500/10"  },
        ].map((s) => {
          const Icon = s.icon;
          return (
            <Card key={s.label} className="py-4">
              <CardContent className="px-4 py-0 flex items-center justify-between">
                <div>
                  <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">{s.label}</p>
                  <p className={`text-xl font-bold tabular-nums ${s.color}`}>{s.value}</p>
                </div>
                <div className={`flex h-8 w-8 items-center justify-center rounded-full ${s.bg}`}>
                  <Icon className={`h-4 w-4 ${s.color}`} />
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Table */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <CardTitle className="text-sm font-semibold">Transactions</CardTitle>
            <span className="text-xs text-muted-foreground">
              {total > 0
                ? `${(page - 1) * PAGE_SIZE + 1}–${Math.min(page * PAGE_SIZE, total)} of ${total}`
                : "No results"}
            </span>
          </div>
        </CardHeader>
        <CardContent className="pt-0 space-y-4">
          <Suspense>
            <TransactionsFilters />
          </Suspense>

          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>Date</TableHead>
                <TableHead>User</TableHead>
                <TableHead>Source</TableHead>
                <TableHead>Plan</TableHead>
                <TableHead className="text-right">Amount</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Provider TX ID</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {transactions.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center text-muted-foreground py-12">
                    No transactions match the current filters.
                  </TableCell>
                </TableRow>
              ) : (
                transactions.map((tx) => (
                  <TableRow key={tx.id}>
                    <TableCell className="text-xs text-muted-foreground whitespace-nowrap">{fmt(tx.created_at)}</TableCell>
                    <TableCell className="font-medium text-sm">{tx.email || tx.user_id.slice(0, 8) + "…"}</TableCell>
                    <TableCell>
                      <span className="text-xs text-muted-foreground">{formatSource(tx.source, tx.platform)}</span>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary" className="text-xs capitalize">{tx.plan_type}</Badge>
                    </TableCell>
                    <TableCell className="text-right font-mono font-medium">
                      {tx.status === "refunded" ? (
                        <span className="text-slate-500">−${tx.amount.toFixed(2)}</span>
                      ) : (
                        <span>${tx.amount.toFixed(2)}</span>
                      )}
                      <span className="ml-1 text-xs text-muted-foreground">{tx.currency}</span>
                    </TableCell>
                    <TableCell>
                      <Badge className={`${STATUS_STYLE[tx.status] ?? "bg-muted"} border text-xs`}>
                        {tx.status}
                      </Badge>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground max-w-[160px] truncate" title={tx.provider_tx_id}>
                      {tx.provider_tx_id || "—"}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {/* Pagination */}
          {total > 0 && (
            <div className="flex items-center justify-between px-1 pt-3 border-t">
              <p className="text-xs text-muted-foreground">
                Page <span className="font-medium text-foreground">{page}</span> of{" "}
                <span className="font-medium text-foreground">{totalPages}</span>
              </p>
              <div className="flex items-center gap-1">
                <Button variant="outline" size="icon" className="h-7 w-7" disabled={page <= 1} asChild={page > 1}>
                  {page > 1 ? <Link href={buildPageUrl(1)}><ChevronsLeft className="h-3.5 w-3.5" /></Link> : <span><ChevronsLeft className="h-3.5 w-3.5" /></span>}
                </Button>
                <Button variant="outline" size="icon" className="h-7 w-7" disabled={page <= 1} asChild={page > 1}>
                  {page > 1 ? <Link href={buildPageUrl(page - 1)}><ChevronLeft className="h-3.5 w-3.5" /></Link> : <span><ChevronLeft className="h-3.5 w-3.5" /></span>}
                </Button>
                {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                  let p: number;
                  if (totalPages <= 5) p = i + 1;
                  else if (page <= 3) p = i + 1;
                  else if (page >= totalPages - 2) p = totalPages - 4 + i;
                  else p = page - 2 + i;
                  return (
                    <Button key={p} variant={p === page ? "default" : "outline"} size="icon" className="h-7 w-7 text-xs" asChild={p !== page}>
                      {p !== page ? <Link href={buildPageUrl(p)}>{p}</Link> : <span>{p}</span>}
                    </Button>
                  );
                })}
                <Button variant="outline" size="icon" className="h-7 w-7" disabled={page >= totalPages} asChild={page < totalPages}>
                  {page < totalPages ? <Link href={buildPageUrl(page + 1)}><ChevronRight className="h-3.5 w-3.5" /></Link> : <span><ChevronRight className="h-3.5 w-3.5" /></span>}
                </Button>
                <Button variant="outline" size="icon" className="h-7 w-7" disabled={page >= totalPages} asChild={page < totalPages}>
                  {page < totalPages ? <Link href={buildPageUrl(totalPages)}><ChevronsRight className="h-3.5 w-3.5" /></Link> : <span><ChevronsRight className="h-3.5 w-3.5" /></span>}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
