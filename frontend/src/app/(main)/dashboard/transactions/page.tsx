import { Suspense } from "react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import {
  AlertCircle, CheckCircle, XCircle, RefreshCw,
  DollarSign, TrendingUp, TrendingDown,
  ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight,
} from "lucide-react";
import { getTransactions } from "@/actions/transactions";
import type { TransactionsParams, TransactionSummary } from "@/actions/transactions";
import { formatSource } from "@/lib/subscriptions/format";
import { TransactionsFilters } from "./_components/transactions-filters";
import { CopyTxId } from "./_components/copy-tx-id";
import { TxRow } from "./_components/tx-row";
import { SortHeader } from "@/components/ui/sort-header";

const PAGE_SIZE = 20;

const STATUS_STYLE: Record<string, string> = {
  success:  "bg-green-500/10 text-green-500 hover:bg-green-500/20 border-green-500/20",
  failed:   "bg-red-500/10 text-red-500 hover:bg-red-500/20 border-red-500/20",
  refunded: "bg-slate-500/10 text-slate-500 hover:bg-slate-500/20 border-slate-500/20",
};

function fmt(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    month: "short", day: "numeric", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

function KPICard({
  title, value, icon: Icon, iconColor, bgColor, sub,
}: {
  title: string;
  value: string | number;
  icon: React.ElementType;
  iconColor: string;
  bgColor: string;
  sub?: React.ReactNode;
}) {
  return (
    <Card className="p-6 hover:shadow-lg transition-all duration-300 hover:-translate-y-0.5">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-sm font-medium text-muted-foreground mb-1">{title}</p>
          <p className="text-3xl font-bold tracking-tight">{value}</p>
          {sub && <div className="mt-2">{sub}</div>}
        </div>
        <div className={`rounded-lg p-3 ${bgColor}`}>
          <Icon className={`h-6 w-6 ${iconColor}`} />
        </div>
      </div>
    </Card>
  );
}

function buildKPIs(s: TransactionSummary) {
  return [
    {
      title: "Total Transactions",
      value: s.total_count.toLocaleString(),
      icon: DollarSign,
      iconColor: "text-primary",
      bgColor: "bg-primary/10",
    },
    {
      title: "Successful",
      value: s.success_count.toLocaleString(),
      icon: CheckCircle,
      iconColor: "text-emerald-500",
      bgColor: "bg-emerald-500/10",
    },
    {
      title: "Failed",
      value: s.failed_count.toLocaleString(),
      icon: XCircle,
      iconColor: "text-red-500",
      bgColor: "bg-red-500/10",
    },
    {
      title: "Refunded",
      value: s.refunded_count.toLocaleString(),
      icon: RefreshCw,
      iconColor: "text-slate-500",
      bgColor: "bg-slate-500/10",
    },
    {
      title: "Total Revenue",
      value: `$${s.total_revenue.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      icon: TrendingUp,
      iconColor: "text-emerald-500",
      bgColor: "bg-emerald-500/10",
    },
    {
      title: "Total Refunds",
      value: `$${s.total_refunded.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`,
      icon: TrendingDown,
      iconColor: "text-slate-500",
      bgColor: "bg-slate-500/10",
    },
  ];
}

interface Props {
  searchParams: Promise<Record<string, string | undefined>>;
}

export default async function TransactionsPage({ searchParams }: Props) {
  const sp = await searchParams;
  const page = Math.max(1, parseInt(sp.page ?? "1", 10) || 1);
  const sort = sp.sort ?? "date_desc"; // default: newest first

  const params: TransactionsParams = {
    page, limit: PAGE_SIZE,
    status: sp.status, source: sp.source, platform: sp.platform,
    search: sp.search, date_from: sp.date_from, date_to: sp.date_to,
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

  const { transactions: rawTxs, summary, total, total_pages: totalPages } = data;

  // Sort by created_at (backend doesn't expose sort param yet)
  const transactions = [...rawTxs].sort((a, b) => {
    const diff = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
    return sort === "date_asc" ? diff : -diff;
  });

  const buildPageUrl = (p: number) => {
    const qs = new URLSearchParams();
    if (sp.status) qs.set("status", sp.status);
    if (sp.source) qs.set("source", sp.source);
    if (sp.platform) qs.set("platform", sp.platform);
    if (sp.search) qs.set("search", sp.search);
    if (sp.date_from) qs.set("date_from", sp.date_from);
    if (sp.date_to) qs.set("date_to", sp.date_to);
    if (sort !== "date_desc") qs.set("sort", sort);
    qs.set("page", String(p));
    return `?${qs.toString()}`;
  };

  const buildSortUrl = (s: string) => {
    const qs = new URLSearchParams();
    if (sp.status) qs.set("status", sp.status);
    if (sp.source) qs.set("source", sp.source);
    if (sp.platform) qs.set("platform", sp.platform);
    if (sp.search) qs.set("search", sp.search);
    if (sp.date_from) qs.set("date_from", sp.date_from);
    if (sp.date_to) qs.set("date_to", sp.date_to);
    if (s !== "date_desc") qs.set("sort", s);
    qs.set("page", "1");
    return `?${qs.toString()}`;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold mb-1">Transaction Reconciliation</h1>
        <p className="text-muted-foreground">Monitor and manage all payment transactions</p>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {buildKPIs(summary).map((kpi) => (
          <KPICard key={kpi.title} {...kpi} />
        ))}
      </div>

      {/* Table card */}
      <Card className="p-6">
        <Suspense>
          <TransactionsFilters />
        </Suspense>

        <div className="rounded-lg border overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/50 hover:bg-muted/50">
                <TableHead className="font-semibold">
                  <SortHeader
                    label="Date"
                    sortKey="date"
                    currentSort={sort}
                    ascHref={buildSortUrl("date_asc")}
                    descHref={buildSortUrl("date_desc")}
                  />
                </TableHead>
                <TableHead className="font-semibold">User Email</TableHead>
                <TableHead className="font-semibold">Source</TableHead>
                <TableHead className="font-semibold">Plan Type</TableHead>
                <TableHead className="font-semibold text-right">Amount</TableHead>
                <TableHead className="font-semibold">Status</TableHead>
                <TableHead className="font-semibold">Provider TX ID</TableHead>
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
                  <TxRow key={tx.id} tx={tx}>
                    <TableCell className="text-muted-foreground whitespace-nowrap">{fmt(tx.created_at)}</TableCell>
                    <TableCell className="font-medium">{tx.email || tx.user_id.slice(0, 8) + "…"}</TableCell>
                    <TableCell>
                      <Badge variant="outline" className="font-medium">{formatSource(tx.source, tx.platform)}</Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground capitalize">{tx.plan_type}</TableCell>
                    <TableCell className="text-right font-semibold">
                      {tx.status === "refunded" ? (
                        <span className="text-slate-500">−${tx.amount.toFixed(2)}</span>
                      ) : (
                        <span>${tx.amount.toFixed(2)}</span>
                      )}
                      <span className="ml-1 text-xs text-muted-foreground">{tx.currency}</span>
                    </TableCell>
                    <TableCell>
                      <Badge className={`${STATUS_STYLE[tx.status] ?? "bg-muted"} border`}>
                        {tx.status.charAt(0).toUpperCase() + tx.status.slice(1)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {tx.provider_tx_id ? <CopyTxId txId={tx.provider_tx_id} /> : <span className="text-muted-foreground">—</span>}
                    </TableCell>
                  </TxRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>

        {/* Pagination */}
        <div className="flex items-center justify-between mt-6">
          <p className="text-sm text-muted-foreground">
            {total > 0
              ? <>Showing <span className="font-medium text-foreground">{(page - 1) * PAGE_SIZE + 1}</span> to{" "}
                  <span className="font-medium text-foreground">{Math.min(page * PAGE_SIZE, total)}</span> of{" "}
                  <span className="font-medium text-foreground">{total}</span> transactions</>
              : "No transactions found"}
          </p>
          {totalPages > 1 && (
            <div className="flex items-center gap-2">
              <Button variant="outline" size="icon" disabled={page <= 1} asChild={page > 1}>
                {page > 1 ? <Link href={buildPageUrl(1)}><ChevronsLeft className="h-4 w-4" /></Link> : <span><ChevronsLeft className="h-4 w-4" /></span>}
              </Button>
              <Button variant="outline" size="icon" disabled={page <= 1} asChild={page > 1}>
                {page > 1 ? <Link href={buildPageUrl(page - 1)}><ChevronLeft className="h-4 w-4" /></Link> : <span><ChevronLeft className="h-4 w-4" /></span>}
              </Button>
              {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                let p: number;
                if (totalPages <= 5) p = i + 1;
                else if (page <= 3) p = i + 1;
                else if (page >= totalPages - 2) p = totalPages - 4 + i;
                else p = page - 2 + i;
                return (
                  <Button key={p} variant={p === page ? "default" : "outline"} size="icon" asChild={p !== page}>
                    {p !== page ? <Link href={buildPageUrl(p)}>{p}</Link> : <span>{p}</span>}
                  </Button>
                );
              })}
              <Button variant="outline" size="icon" disabled={page >= totalPages} asChild={page < totalPages}>
                {page < totalPages ? <Link href={buildPageUrl(page + 1)}><ChevronRight className="h-4 w-4" /></Link> : <span><ChevronRight className="h-4 w-4" /></span>}
              </Button>
              <Button variant="outline" size="icon" disabled={page >= totalPages} asChild={page < totalPages}>
                {page < totalPages ? <Link href={buildPageUrl(totalPages)}><ChevronsRight className="h-4 w-4" /></Link> : <span><ChevronsRight className="h-4 w-4" /></span>}
              </Button>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
