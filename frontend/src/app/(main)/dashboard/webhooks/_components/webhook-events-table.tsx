"use client";

import React, { useState, useMemo } from "react";
import { Search, Copy, Eye, RotateCw, ChevronLeft, ChevronRight, Check, X, ChevronsLeft, ChevronsRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { replayWebhook } from "@/actions/revenue-ops";
import type { WebhookEvent, WebhookSummary } from "@/actions/webhooks";

const providerColor: Record<string, string> = {
  stripe: "bg-purple-500/10 text-purple-500 border-purple-500/20",
  apple:  "bg-gray-500/10 text-gray-400 border-gray-500/20",
  google: "bg-blue-500/10 text-blue-500 border-blue-500/20",
};

const statusColor: Record<string, string> = {
  pending:   "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
  processed: "bg-green-500/10 text-green-500 border-green-500/20",
};

function fmt(iso: string) {
  return new Date(iso).toLocaleString("en-US", {
    month: "short", day: "numeric", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

interface Props {
  webhooks: WebhookEvent[];
  summary: WebhookSummary;
  total: number;
  page: number;
  totalPages: number;
  initialProvider?: string;
  initialStatus?: string;
  initialSearch?: string;
}

export function WebhookEventsTable({
  webhooks, summary, total, page, totalPages,
  initialProvider = "all", initialStatus = "all", initialSearch = "",
}: Props) {
  const [search, setSearch] = useState(initialSearch);
  const [provider, setProvider] = useState(initialProvider);
  const [status, setStatus] = useState(initialStatus);
  const [selected, setSelected] = useState<WebhookEvent | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [replayingId, setReplayingId] = useState<string | null>(null);
  const [replayedIds, setReplayedIds] = useState<Set<string>>(new Set());

  const filtered = useMemo(() => {
    return webhooks.filter((e) => {
      const matchProvider = provider === "all" || e.provider.toLowerCase() === provider.toLowerCase();
      const matchStatus = status === "all"
        || (status === "pending" && !e.processed)
        || (status === "processed" && e.processed);
      const matchSearch = !search
        || e.event_id.toLowerCase().includes(search.toLowerCase())
        || e.event_type.toLowerCase().includes(search.toLowerCase());
      return matchProvider && matchStatus && matchSearch;
    });
  }, [webhooks, provider, status, search]);

  const buildUrl = (p: number) => {
    const qs = new URLSearchParams();
    if (provider !== "all") qs.set("provider", provider);
    if (status !== "all") qs.set("status", status);
    if (search) qs.set("search", search);
    qs.set("page", String(p));
    return `?${qs.toString()}`;
  };

  function handleCopy(text: string) {
    navigator.clipboard.writeText(text);
    setCopiedId(text);
    setTimeout(() => setCopiedId(null), 2000);
  }

  async function handleReplay(id: string) {
    setReplayingId(id);
    try {
      await replayWebhook(id);
      setReplayedIds((prev) => new Set([...prev, id]));
    } finally {
      setReplayingId(null);
    }
  }

  const failedCount = Math.max(0, summary.total - summary.pending - summary.processed);

  return (
    <>
      {/* Stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-5 pb-4">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">Total</p>
            <p className="text-3xl font-bold tabular-nums">{summary.total.toLocaleString()}</p>
            <p className="text-xs text-muted-foreground mt-1">All events</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5 pb-4">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">Pending</p>
            <p className="text-3xl font-bold tabular-nums text-yellow-500">{summary.pending.toLocaleString()}</p>
            <p className="text-xs text-muted-foreground mt-1">Awaiting processing</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5 pb-4">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">Processed</p>
            <p className="text-3xl font-bold tabular-nums text-green-500">{summary.processed.toLocaleString()}</p>
            <p className="text-xs text-muted-foreground mt-1">Successfully handled</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-5 pb-4">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-widest mb-1">Failed</p>
            <p className="text-3xl font-bold tabular-nums text-red-500">{failedCount.toLocaleString()}</p>
            <p className="text-xs text-muted-foreground mt-1">Requires attention</p>
          </CardContent>
        </Card>
      </div>

      {/* Table card */}
      <Card>
        <CardHeader>
          <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between">
              <CardTitle className="text-base">Event Log</CardTitle>
            </div>
            <div className="flex flex-col sm:flex-row gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
                <Input
                  placeholder="Search event ID or type…"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="pl-9 pr-8 w-full"
                />
                {search && (
                  <button
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    onClick={() => setSearch("")}
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
              <Select value={provider} onValueChange={setProvider}>
                <SelectTrigger className="w-full sm:w-[150px] shrink-0">
                  <SelectValue placeholder="Provider" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Providers</SelectItem>
                  <SelectItem value="stripe">Stripe</SelectItem>
                  <SelectItem value="apple">Apple</SelectItem>
                  <SelectItem value="google">Google</SelectItem>
                </SelectContent>
              </Select>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger className="w-full sm:w-[140px] shrink-0">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Status</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="processed">Processed</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow className="bg-muted/40 hover:bg-muted/40">
                  <TableHead className="text-xs font-semibold h-9">Provider</TableHead>
                  <TableHead className="text-xs font-semibold h-9">Event Type</TableHead>
                  <TableHead className="text-xs font-semibold h-9">Event ID</TableHead>
                  <TableHead className="text-xs font-semibold h-9">Received</TableHead>
                  <TableHead className="text-xs font-semibold h-9">Status</TableHead>
                  <TableHead className="text-xs font-semibold h-9 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-12 text-muted-foreground text-sm">
                      No events match your filters
                    </TableCell>
                  </TableRow>
                ) : (
                  filtered.map((event) => {
                    const statusKey = event.processed ? "processed" : "pending";
                    return (
                      <TableRow key={event.id} className="hover:bg-muted/30">
                        <TableCell>
                          <Badge variant="outline" className={`${providerColor[event.provider.toLowerCase()] ?? "bg-muted"} capitalize text-xs`}>
                            {event.provider}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <code className="text-xs bg-muted px-2 py-0.5 rounded font-mono">{event.event_type}</code>
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center gap-1.5">
                            <span className="font-mono text-xs text-muted-foreground max-w-[140px] truncate">{event.event_id || event.id.slice(0, 8)}</span>
                            <button
                              className="text-muted-foreground hover:text-foreground transition-colors"
                              onClick={() => handleCopy(event.event_id || event.id)}
                            >
                              {copiedId === (event.event_id || event.id)
                                ? <Check className="h-3 w-3 text-green-500" />
                                : <Copy className="h-3 w-3" />}
                            </button>
                          </div>
                        </TableCell>
                        <TableCell className="text-xs text-muted-foreground whitespace-nowrap" suppressHydrationWarning>{fmt(event.created_at)}</TableCell>
                        <TableCell>
                          <Badge variant="outline" className={`${statusColor[statusKey]} text-xs`}>
                            <span className={`inline-block w-1.5 h-1.5 rounded-full mr-1.5 ${statusKey === "pending" ? "bg-yellow-500 animate-pulse" : "bg-green-500"}`} />
                            {statusKey}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setSelected(event)}>
                              <Eye className="h-3.5 w-3.5" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-7 w-7"
                              disabled={replayingId === event.id || replayedIds.has(event.id)}
                              onClick={() => handleReplay(event.id)}
                            >
                              {replayedIds.has(event.id)
                                ? <Check className="h-3.5 w-3.5 text-green-500" />
                                : <RotateCw className={`h-3.5 w-3.5 ${replayingId === event.id ? "animate-spin" : ""}`} />}
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between mt-5">
              <p className="text-sm text-muted-foreground">
                Showing{" "}
                <span className="font-medium text-foreground">{(page - 1) * 20 + 1}</span>–
                <span className="font-medium text-foreground">{Math.min(page * 20, total)}</span> of{" "}
                <span className="font-medium text-foreground">{total.toLocaleString()}</span>
              </p>
              <div className="flex items-center gap-1.5">
                <Button variant="outline" size="icon" className="h-8 w-8" disabled={page <= 1} asChild={page > 1}>
                  {page > 1 ? <a href={buildUrl(1)}><ChevronsLeft className="h-4 w-4" /></a> : <span><ChevronsLeft className="h-4 w-4" /></span>}
                </Button>
                <Button variant="outline" size="icon" className="h-8 w-8" disabled={page <= 1} asChild={page > 1}>
                  {page > 1 ? <a href={buildUrl(page - 1)}><ChevronLeft className="h-4 w-4" /></a> : <span><ChevronLeft className="h-4 w-4" /></span>}
                </Button>
                <span className="text-sm px-2 text-muted-foreground tabular-nums">Page {page} of {totalPages}</span>
                <Button variant="outline" size="icon" className="h-8 w-8" disabled={page >= totalPages} asChild={page < totalPages}>
                  {page < totalPages ? <a href={buildUrl(page + 1)}><ChevronRight className="h-4 w-4" /></a> : <span><ChevronRight className="h-4 w-4" /></span>}
                </Button>
                <Button variant="outline" size="icon" className="h-8 w-8" disabled={page >= totalPages} asChild={page < totalPages}>
                  {page < totalPages ? <a href={buildUrl(totalPages)}><ChevronsRight className="h-4 w-4" /></a> : <span><ChevronsRight className="h-4 w-4" /></span>}
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Detail sheet */}
      <Sheet open={!!selected} onOpenChange={() => setSelected(null)}>
        <SheetContent className="sm:max-w-lg overflow-y-auto">
          {selected && (
            <>
              <SheetHeader className="pb-5 border-b mb-5">
                <SheetTitle className="text-base font-semibold">Webhook Event</SheetTitle>
                <SheetDescription className="font-mono text-xs truncate">{selected.id}</SheetDescription>
              </SheetHeader>
              <div className="space-y-5 pb-6">
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className={`${providerColor[selected.provider.toLowerCase()] ?? "bg-muted"} capitalize font-semibold px-3 py-0.5 text-xs rounded-full`}>
                    {selected.provider}
                  </Badge>
                  <Badge variant="outline" className={`${statusColor[selected.processed ? "processed" : "pending"]} text-xs rounded-full`}>
                    <span className={`inline-block w-1.5 h-1.5 rounded-full mr-1.5 ${!selected.processed ? "bg-yellow-500 animate-pulse" : "bg-green-500"}`} />
                    {selected.processed ? "processed" : "pending"}
                  </Badge>
                </div>

                <div className="flex flex-col gap-1">
                  <span className="text-[10px] uppercase tracking-widest font-semibold text-muted-foreground/70">Event Type</span>
                  <code className="text-xs bg-muted px-3 py-2 rounded block font-mono">{selected.event_type}</code>
                </div>

                <div className="flex flex-col gap-1">
                  <span className="text-[10px] uppercase tracking-widest font-semibold text-muted-foreground/70">Received</span>
                  <span className="text-sm" suppressHydrationWarning>{fmt(selected.created_at)}</span>
                </div>

                {selected.processed_at && (
                  <div className="flex flex-col gap-1">
                    <span className="text-[10px] uppercase tracking-widest font-semibold text-muted-foreground/70">Processed At</span>
                    <span className="text-sm" suppressHydrationWarning>{fmt(selected.processed_at)}</span>
                  </div>
                )}

                <div className="flex flex-col gap-1">
                  <span className="text-[10px] uppercase tracking-widest font-semibold text-muted-foreground/70">Event ID</span>
                  <div className="flex items-center gap-2 bg-muted rounded p-2">
                    <code className="font-mono text-xs flex-1 break-all">{selected.event_id || "—"}</code>
                    {selected.event_id && (
                      <Button variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={() => handleCopy(selected.event_id)}>
                        {copiedId === selected.event_id ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
                      </Button>
                    )}
                  </div>
                </div>

                <div className="flex flex-col gap-1">
                  <span className="text-[10px] uppercase tracking-widest font-semibold text-muted-foreground/70">Internal ID</span>
                  <div className="flex items-center gap-2 bg-muted rounded p-2">
                    <code className="font-mono text-xs flex-1 break-all">{selected.id}</code>
                    <Button variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={() => handleCopy(selected.id)}>
                      {copiedId === selected.id ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
                    </Button>
                  </div>
                </div>

                <Button
                  className="w-full"
                  variant={replayedIds.has(selected.id) ? "outline" : "default"}
                  disabled={replayingId === selected.id || replayedIds.has(selected.id)}
                  onClick={() => handleReplay(selected.id)}
                >
                  {replayedIds.has(selected.id) ? (
                    <><Check className="h-4 w-4 mr-2 text-green-500" /> Queued for replay</>
                  ) : (
                    <><RotateCw className={`h-4 w-4 mr-2 ${replayingId === selected.id ? "animate-spin" : ""}`} /> Replay Event</>
                  )}
                </Button>
              </div>
            </>
          )}
        </SheetContent>
      </Sheet>
    </>
  );
}
