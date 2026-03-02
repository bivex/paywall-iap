"use client";

import { useState } from "react";
import { AlertTriangle, RefreshCw, ChevronDown, ChevronUp, Wifi } from "lucide-react";
import { getFailedRequests } from "@/lib/network-monitor.client";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string; componentStack?: string };
  reset: () => void;
}) {
  const [stackExpanded, setStackExpanded] = useState(true); // open by default
  const isDev = process.env.NODE_ENV === "development";
  const lines = parseStack(error.stack);
  const appLines = lines.filter((l) => l.isApp);
  const componentPath = extractComponentPath(lines);
  const failedRequests = isDev ? getFailedRequests() : [];

  // Use real componentStack if available (client-side errors),
  // otherwise show synthesised path from JS stack (server-side errors)
  const compStack = error.componentStack?.trim() ?? null;

  return (
    <div className="flex flex-col gap-4 py-16 px-4 max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-500/10">
          <AlertTriangle className="h-5 w-5 text-red-500" />
        </div>
        <div>
          <h2 className="text-base font-semibold text-foreground">Something went wrong</h2>
          <p className="text-sm text-muted-foreground">{error.name}</p>
        </div>
        {error.digest && (
          <code className="ml-auto text-[10px] text-muted-foreground bg-muted px-2 py-0.5 rounded">
            digest:{error.digest}
          </code>
        )}
      </div>

      {/* Error message */}
      <div className="rounded-lg bg-red-500/5 border border-red-500/20 px-4 py-3">
        <p className="text-sm text-red-700 dark:text-red-400 font-mono break-all">{error.message}</p>
      </div>

      {isDev && (
        <>
          {/* Component path — always shown */}
          <div className="rounded-lg border bg-muted/30 overflow-hidden">
            <div className="px-4 py-2 text-xs font-medium text-muted-foreground border-b">
              📍 Component path
              {!compStack && (
                <span className="ml-2 text-slate-400">(server error — synthesised from stack)</span>
              )}
            </div>
            {compStack ? (
              <pre className="bg-slate-950 px-4 py-3 text-emerald-400 text-xs overflow-x-auto max-h-48 overflow-y-auto whitespace-pre-wrap">
                {compStack}
              </pre>
            ) : componentPath.length > 0 ? (
              <div className="bg-slate-950 px-4 py-3 flex flex-wrap gap-1.5">
                {componentPath.map((seg, i) => (
                  <span key={i} className="flex items-center gap-1">
                    {i > 0 && <span className="text-slate-600">›</span>}
                    <code className={`text-xs px-1.5 py-0.5 rounded ${seg.isPage ? "bg-orange-900/40 text-orange-300 font-bold" : "bg-slate-800 text-slate-300"}`}>
                      {seg.label}
                    </code>
                    {seg.line && <span className="text-slate-600 text-[10px]">:{seg.line}</span>}
                  </span>
                ))}
              </div>
            ) : (
              <div className="bg-slate-950 px-4 py-3 text-slate-500 text-xs">No component info available</div>
            )}
          </div>

          {/* Network */}
          <div className="rounded-lg border bg-muted/30 overflow-hidden">
            <div className="flex items-center gap-2 px-4 py-2 text-xs font-medium text-muted-foreground border-b">
              <Wifi className="h-3.5 w-3.5" />
              Network
              {failedRequests.length > 0 ? (
                <span className="text-amber-500 font-semibold">
                  {failedRequests.length} failed request{failedRequests.length > 1 ? "s" : ""}
                </span>
              ) : (
                <span className="text-slate-500">no client-side failures detected</span>
              )}
              <span className="ml-auto text-[10px] text-slate-600">server actions not monitored</span>
            </div>
            {failedRequests.length > 0 ? (
              <div className="bg-slate-950 divide-y divide-slate-800">
                {failedRequests.map((r, i) => (
                  <div key={i} className="px-4 py-2.5 font-mono">
                    <div className="flex items-center gap-2 text-xs">
                      <span className={`font-bold ${r.status >= 500 ? "text-red-400" : r.status >= 400 ? "text-amber-400" : r.status === 0 ? "text-red-400" : "text-slate-400"}`}>
                        {r.status || "NET ERR"}
                      </span>
                      <span className="text-slate-300 truncate">{r.method} {r.url}</span>
                      <span className="ml-auto text-slate-600 text-[10px] shrink-0">{r.time}</span>
                    </div>
                    {r.body && (
                      <pre className="mt-1 text-[10px] text-slate-500 truncate">{r.body}</pre>
                    )}
                  </div>
                ))}
              </div>
            ) : (
              <div className="bg-slate-950 px-4 py-3 text-slate-600 text-xs">
                Client-side fetch calls will appear here if they fail.
              </div>
            )}
          </div>

          {/* JS Stack — open by default */}
          {lines.length > 0 && (
            <div className="rounded-lg border bg-muted/30 overflow-hidden">
              <button
                className="w-full flex items-center justify-between px-4 py-2.5 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors"
                onClick={() => setStackExpanded((v) => !v)}
              >
                <span>
                  JS stack trace
                  {appLines.length > 0 && (
                    <span className="ml-2 text-orange-500 font-semibold">
                      {appLines.length} frame{appLines.length > 1 ? "s" : ""} in your code
                    </span>
                  )}
                </span>
                {stackExpanded ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
              </button>

              {stackExpanded && (
                <div className="bg-slate-950 px-4 py-3 overflow-x-auto max-h-80 overflow-y-auto">
                  {lines.map((line, i) => (
                    <div key={i} className="leading-5">
                      {line.isApp ? (
                        <span className="text-orange-400 font-semibold text-xs">{line.raw}</span>
                      ) : (
                        <span className="text-slate-500 text-[11px]">{line.raw}</span>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </>
      )}

      <button
        onClick={reset}
        className="flex items-center gap-2 self-start rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
      >
        <RefreshCw className="h-3.5 w-3.5" />
        Try again
      </button>
    </div>
  );
}

function parseStack(stack?: string) {
  if (!stack) return [];
  return stack
    .split("\n")
    .slice(1)
    .map((raw) => ({
      raw: raw.trim(),
      isApp: /\bsrc\/|\/app\/|\/actions\/|\/components\//.test(raw) && !raw.includes("node_modules"),
    }));
}

/** Extract component-like segments from JS stack for server-side errors. */
function extractComponentPath(lines: ReturnType<typeof parseStack>) {
  const seen = new Set<string>();
  const segments: { label: string; line: string | null; isPage: boolean }[] = [];

  for (const { raw, isApp } of lines) {
    if (!isApp) continue;
    // Match "at FunctionName (path/to/file.tsx:line:col)"
    const match = raw.match(/^at\s+(\S+)\s+\((.+?):(\d+):\d+\)/);
    if (!match) continue;
    const [, fn, , lineNum] = match;
    // Skip internal helpers
    if (/^(Object|Module|Promise|async|eval|<anonymous>)/.test(fn)) continue;
    if (seen.has(fn)) continue;
    seen.add(fn);
    segments.push({
      label: fn,
      line: lineNum ?? null,
      isPage: /Page|Layout|Error|Loading/.test(fn),
    });
    if (segments.length >= 6) break;
  }

  return segments;
}
