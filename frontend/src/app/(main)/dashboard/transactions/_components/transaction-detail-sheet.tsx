"use client";

import React, { cloneElement, isValidElement, useState, useTransition } from "react";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Loader2, ExternalLink, Copy, Check, User, CreditCard } from "lucide-react";
import { getTransactionDetail } from "@/actions/transactions";
import type { TransactionDetail } from "@/actions/transactions";
import { formatSource, formatPlanType } from "@/lib/subscriptions/format";

const txStatusClass: Record<string, string> = {
  success:  "bg-green-500/10 text-green-500 border-green-500/20",
  failed:   "bg-red-500/10 text-red-500 border-red-500/20",
  refunded: "bg-blue-500/10 text-blue-500 border-blue-500/20",
};

const subStatusClass: Record<string, string> = {
  active:    "bg-green-500/10 text-green-500 border-green-500/20",
  grace:     "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
  cancelled: "bg-orange-500/10 text-orange-500 border-orange-500/20",
  expired:   "bg-red-500/10 text-red-500 border-red-500/20",
};

function fmt(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    year: "numeric", month: "short", day: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

function CopyBtn({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      className="ml-1 inline-flex items-center text-muted-foreground hover:text-foreground transition-colors"
      onClick={(e) => {
        e.stopPropagation();
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
      }}
    >
      {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
    </button>
  );
}

function Field({ label, value, mono }: { label: string; value: React.ReactNode; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-1">
      <span className="text-[10px] font-semibold text-muted-foreground/70 uppercase tracking-widest">{label}</span>
      <span className={`text-sm ${mono ? "font-mono text-xs break-all text-foreground/80" : "font-medium"}`}>{value}</span>
    </div>
  );
}

function SectionCard({ icon: Icon, title, children }: { icon: React.ElementType; title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border bg-muted/30 overflow-hidden">
      <div className="flex items-center gap-2 px-4 py-2.5 border-b bg-muted/40">
        <Icon className="h-3.5 w-3.5 text-muted-foreground" />
        <span className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">{title}</span>
      </div>
      <div className="p-4">{children}</div>
    </div>
  );
}

interface Props {
  transactionId: string;
  trigger: React.ReactNode;
}

export function TransactionDetailSheet({ transactionId, trigger }: Props) {
  const [open, setOpen] = useState(false);
  const [detail, setDetail] = useState<TransactionDetail | null>(null);
  const [isPending, startTransition] = useTransition();

  function handleOpen() {
    setOpen(true);
    if (!detail) {
      startTransition(async () => {
        const data = await getTransactionDetail(transactionId);
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
        <SheetContent className="w-full sm:max-w-lg overflow-y-auto">
          <SheetHeader className="pb-5 border-b mb-5">
            <SheetTitle className="text-base font-semibold">Transaction Detail</SheetTitle>
          </SheetHeader>

          {isPending && (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          )}

          {!isPending && !detail && (
            <p className="text-sm text-muted-foreground py-8 text-center">Failed to load transaction.</p>
          )}

          {!isPending && detail && (
            <div className="space-y-5 pb-6">
              {/* Status + amount hero */}
              <div className="rounded-lg border bg-muted/30 p-4 flex items-center justify-between">
                <div>
                  <p className="text-2xl font-bold tabular-nums">
                    {detail.status === "refunded" ? "−" : ""}${detail.amount.toFixed(2)}
                    <span className="ml-1.5 text-sm font-normal text-muted-foreground">{detail.currency.toUpperCase()}</span>
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5">{fmt(detail.created_at)}</p>
                </div>
                <Badge variant="outline" className={`${txStatusClass[detail.status] ?? "bg-muted"} capitalize font-semibold px-3 py-1 text-xs rounded-full`}>
                  {detail.status}
                </Badge>
              </div>

              {/* IDs */}
              <div className="rounded-lg border bg-muted/30 p-4 space-y-3">
                <div className="flex flex-col gap-1">
                  <span className="text-[10px] font-semibold text-muted-foreground/70 uppercase tracking-widest">Transaction ID</span>
                  <span className="font-mono text-xs break-all text-foreground/80 flex items-center">
                    {detail.id}
                    <CopyBtn text={detail.id} />
                  </span>
                </div>
                {detail.provider_tx_id && (
                  <div className="flex flex-col gap-1">
                    <span className="text-[10px] font-semibold text-muted-foreground/70 uppercase tracking-widest">Provider Tx ID</span>
                    <span className="font-mono text-xs break-all text-foreground/80 flex items-center">
                      {detail.provider_tx_id}
                      <CopyBtn text={detail.provider_tx_id} />
                    </span>
                  </div>
                )}
                {detail.receipt_hash && (
                  <div className="flex flex-col gap-1">
                    <span className="text-[10px] font-semibold text-muted-foreground/70 uppercase tracking-widest">Receipt Hash</span>
                    <span className="font-mono text-xs break-all text-foreground/80 flex items-center">
                      {detail.receipt_hash}
                      <CopyBtn text={detail.receipt_hash} />
                    </span>
                  </div>
                )}
              </div>

              <Separator />

              {/* User */}
              <SectionCard icon={User} title="User">
                <div className="space-y-3">
                  <div className="grid grid-cols-2 gap-3">
                    <Field label="Email" value={detail.user.email || "—"} />
                    <Field label="LTV" value={
                      <span className="text-base font-bold tabular-nums">${detail.user.ltv.toFixed(2)}</span>
                    } />
                  </div>
                  <Field label="User ID" value={
                    <span className="flex items-center">
                      {detail.user.id}
                      <CopyBtn text={detail.user.id} />
                    </span>
                  } mono />
                  <Field label="Registered" value={fmt(detail.user.created_at)} />
                  <Button variant="outline" size="sm" asChild className="w-full mt-1">
                    <a href={`/dashboard/users/${detail.user.id}`}>
                      <ExternalLink className="mr-2 h-3.5 w-3.5" />
                      View User Profile
                    </a>
                  </Button>
                </div>
              </SectionCard>

              {/* Subscription */}
              <SectionCard icon={CreditCard} title="Subscription">
                <div className="space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant="outline" className={`${subStatusClass[detail.subscription.status] ?? "bg-muted"} capitalize font-semibold px-3 py-0.5 text-xs rounded-full`}>
                      {detail.subscription.status}
                    </Badge>
                    <span className="rounded-full border bg-muted/40 px-3 py-0.5 text-xs text-muted-foreground font-medium">
                      {formatPlanType(detail.subscription.plan_type)}
                    </span>
                    <span className="rounded-full border bg-muted/40 px-3 py-0.5 text-xs text-muted-foreground font-medium">
                      {formatSource(detail.subscription.source, detail.subscription.platform)}
                    </span>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <Field label="Created" value={fmt(detail.subscription.created_at)} />
                    <Field label="Expires" value={fmt(detail.subscription.expires_at)} />
                  </div>
                  <Field label="Subscription ID" value={
                    <span className="flex items-center">
                      {detail.subscription.id}
                      <CopyBtn text={detail.subscription.id} />
                    </span>
                  } mono />
                  <Button variant="outline" size="sm" asChild className="w-full mt-1">
                    <a href={`/dashboard/subscriptions?search=${detail.user.email}`}>
                      <ExternalLink className="mr-2 h-3.5 w-3.5" />
                      View Subscription
                    </a>
                  </Button>
                </div>
              </SectionCard>
            </div>
          )}
        </SheetContent>
      </Sheet>
    </>
  );
}
