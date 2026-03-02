"use client";

import React, { cloneElement, isValidElement, useState, useTransition } from "react";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Loader2, ExternalLink } from "lucide-react";
import { getSubscriptionDetail } from "@/actions/subscriptions";
import type { SubscriptionDetail } from "@/actions/subscriptions";
import { formatSource, formatPlanType } from "@/lib/subscriptions/format";

const statusClass: Record<string, string> = {
  active:    "bg-green-500/10 text-green-500 border-green-500/20",
  grace:     "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
  cancelled: "bg-orange-500/10 text-orange-500 border-orange-500/20",
  expired:   "bg-red-500/10 text-red-500 border-red-500/20",
};

const txStatusClass: Record<string, string> = {
  success:  "bg-green-500/10 text-green-500",
  pending:  "bg-yellow-500/10 text-yellow-500",
  failed:   "bg-red-500/10 text-red-500",
  refunded: "bg-blue-500/10 text-blue-500",
};

function fmt(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    year: "numeric", month: "short", day: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

function Field({ label, value, mono }: { label: string; value: React.ReactNode; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-[10px] font-semibold text-muted-foreground/70 uppercase tracking-widest">{label}</span>
      <span className={`text-sm ${mono ? "font-mono text-xs break-all text-foreground/80" : "font-medium"}`}>{value}</span>
    </div>
  );
}

function InfoCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="rounded-lg border bg-muted/30 p-4">
      {children}
    </div>
  );
}

interface Props {
  subscriptionId: string;
  trigger: React.ReactNode;
}

export function SubscriptionDetailSheet({ subscriptionId, trigger }: Props) {
  const [open, setOpen] = useState(false);
  const [detail, setDetail] = useState<SubscriptionDetail | null>(null);
  const [isPending, startTransition] = useTransition();

  function handleOpen() {
    setOpen(true);
    if (!detail) {
      startTransition(async () => {
        const data = await getSubscriptionDetail(subscriptionId);
        setDetail(data);
      });
    }
  }

  return (
    <>
      {isValidElement(trigger)
        ? cloneElement(trigger as React.ReactElement<{ onClick?: () => void }>, { onClick: handleOpen })
        : <div onClick={handleOpen} className="cursor-pointer">{trigger}</div>
      }

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent className="w-full sm:max-w-xl overflow-y-auto">
          <SheetHeader className="pb-5 border-b mb-5">
            <SheetTitle className="text-base font-semibold">Subscription Detail</SheetTitle>
          </SheetHeader>

          {isPending && (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          )}

          {!isPending && !detail && (
            <p className="text-sm text-muted-foreground py-8 text-center">Failed to load subscription.</p>
          )}

          {!isPending && detail && (
            <div className="space-y-5 pb-6">
              {/* Header: status + pills */}
              <div className="flex flex-wrap items-center gap-2">
                <Badge variant="outline" className={`${statusClass[detail.status] ?? "bg-muted text-muted-foreground"} capitalize font-semibold px-3 py-0.5 text-xs rounded-full`}>
                  {detail.status}
                </Badge>
                <span className="rounded-full border bg-muted/40 px-3 py-0.5 text-xs text-muted-foreground font-medium">
                  {formatPlanType(detail.plan_type)}
                </span>
                <span className="rounded-full border bg-muted/40 px-3 py-0.5 text-xs text-muted-foreground font-medium">
                  {formatSource(detail.source, detail.platform)}
                </span>
              </div>

              {/* User info card */}
              <InfoCard>
                <div className="grid grid-cols-2 gap-4">
                  <Field label="Email" value={detail.email || <span className="text-muted-foreground">—</span>} />
                  <Field label="LTV" value={
                    <span className="text-lg font-bold tabular-nums">${detail.ltv.toFixed(2)}</span>
                  } />
                </div>
              </InfoCard>

              {/* IDs card */}
              <InfoCard>
                <div className="space-y-3">
                  <Field label="Subscription ID" value={detail.id} mono />
                  <Field label="User ID" value={detail.user_id} mono />
                </div>
              </InfoCard>

              {/* Dates card */}
              <InfoCard>
                <div className="grid grid-cols-3 gap-4">
                  <Field label="Created" value={fmt(detail.created_at)} />
                  <Field label="Updated" value={fmt(detail.updated_at)} />
                  <Field label="Expires" value={fmt(detail.expires_at)} />
                </div>
              </InfoCard>

              <Button variant="outline" size="sm" asChild className="w-full">
                <a href={`/dashboard/users/${detail.user_id}`}>
                  <ExternalLink className="mr-2 h-3.5 w-3.5" />
                  View User Profile
                </a>
              </Button>

              <Separator />

              {/* Transactions */}
              <div>
                <div className="flex items-center gap-2 mb-3">
                  <p className="text-sm font-semibold">Transactions</p>
                  <span className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground font-medium tabular-nums">
                    {detail.transactions.length}
                  </span>
                </div>
                {detail.transactions.length === 0 ? (
                  <div className="rounded-lg border border-dashed py-8 text-center">
                    <p className="text-sm text-muted-foreground">No transactions found.</p>
                  </div>
                ) : (
                  <div className="rounded-lg border overflow-hidden">
                    <Table>
                      <TableHeader>
                        <TableRow className="hover:bg-transparent bg-muted/40">
                          <TableHead className="text-xs h-8 py-2">Provider</TableHead>
                          <TableHead className="text-xs h-8 py-2">Tx ID</TableHead>
                          <TableHead className="text-xs h-8 py-2 text-right">Amount</TableHead>
                          <TableHead className="text-xs h-8 py-2">Status</TableHead>
                          <TableHead className="text-xs h-8 py-2">Date</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {detail.transactions.map((tx) => (
                          <TableRow key={tx.id} className="text-xs">
                            <TableCell className="capitalize py-2.5 font-medium">{tx.provider}</TableCell>
                            <TableCell className="font-mono max-w-[90px] truncate py-2.5 text-muted-foreground">{tx.provider_tx_id}</TableCell>
                            <TableCell className="py-2.5 text-right font-semibold tabular-nums">
                              {tx.amount.toFixed(2)} <span className="text-muted-foreground text-[10px]">{tx.currency.toUpperCase()}</span>
                            </TableCell>
                            <TableCell className="py-2.5">
                              <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-semibold ${txStatusClass[tx.status] ?? "bg-muted text-muted-foreground"}`}>
                                {tx.status}
                              </span>
                            </TableCell>
                            <TableCell className="py-2.5 whitespace-nowrap text-muted-foreground">{fmt(tx.created_at)}</TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </div>
            </div>
          )}
        </SheetContent>
      </Sheet>
    </>
  );
}
