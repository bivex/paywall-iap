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
  active:    "bg-green-100 text-green-800",
  grace:     "bg-yellow-100 text-yellow-800",
  cancelled: "bg-orange-100 text-orange-800",
  expired:   "bg-red-100 text-red-800",
};

function fmt(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    year: "numeric", month: "short", day: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

function Field({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-0.5">
      <span className="text-xs text-muted-foreground uppercase tracking-wide">{label}</span>
      <span className="text-sm font-medium">{value}</span>
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
          <SheetHeader className="pb-4">
            <SheetTitle className="text-base">Subscription Detail</SheetTitle>
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
            <div className="space-y-6">
              {/* Status + plan */}
              <div className="flex items-center gap-3">
                <Badge className={`${statusClass[detail.status] ?? "bg-muted"} text-sm px-3 py-1`}>
                  {detail.status}
                </Badge>
                <span className="text-sm text-muted-foreground">{formatPlanType(detail.plan_type)}</span>
                <span className="text-sm text-muted-foreground">{formatSource(detail.source, detail.platform)}</span>
              </div>

              {/* Core fields */}
              <div className="grid grid-cols-2 gap-4">
                <Field label="Subscription ID" value={
                  <span className="font-mono text-xs break-all">{detail.id}</span>
                } />
                <Field label="User ID" value={
                  <span className="font-mono text-xs break-all">{detail.user_id}</span>
                } />
                <Field label="Email" value={detail.email || "—"} />
                <Field label="LTV" value={`$${detail.ltv.toFixed(2)}`} />
                <Field label="Created" value={fmt(detail.created_at)} />
                <Field label="Updated" value={fmt(detail.updated_at)} />
                <Field label="Expires" value={fmt(detail.expires_at)} />
              </div>

              <div className="flex gap-2">
                <Button variant="outline" size="sm" asChild>
                  <a href={`/dashboard/users/${detail.user_id}`}>
                    View User <ExternalLink className="ml-1.5 h-3 w-3" />
                  </a>
                </Button>
              </div>

              <Separator />

              {/* Transactions */}
              <div>
                <p className="text-sm font-semibold mb-3">
                  Transactions ({detail.transactions.length})
                </p>
                {detail.transactions.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No transactions found.</p>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow className="hover:bg-transparent">
                        <TableHead className="text-xs">Provider</TableHead>
                        <TableHead className="text-xs">Tx ID</TableHead>
                        <TableHead className="text-xs">Amount</TableHead>
                        <TableHead className="text-xs">Status</TableHead>
                        <TableHead className="text-xs">Date</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {detail.transactions.map((tx) => (
                        <TableRow key={tx.id}>
                          <TableCell className="text-xs capitalize">{tx.provider}</TableCell>
                          <TableCell className="font-mono text-xs max-w-[100px] truncate">{tx.provider_tx_id}</TableCell>
                          <TableCell className="text-xs">{tx.amount.toFixed(2)} {tx.currency.toUpperCase()}</TableCell>
                          <TableCell className="text-xs">{tx.status}</TableCell>
                          <TableCell className="text-xs whitespace-nowrap">{fmt(tx.created_at)}</TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </div>
            </div>
          )}
        </SheetContent>
      </Sheet>
    </>
  );
}
