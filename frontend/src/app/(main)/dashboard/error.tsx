"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, RefreshCw, ChevronDown, ChevronUp } from "lucide-react";
import { logError } from "@/lib/debug-error";

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string; componentStack?: string };
  reset: () => void;
}) {
  const [stackExpanded, setStackExpanded] = useState(false);
  const [compExpanded, setCompExpanded] = useState(true);
  const isDev = process.env.NODE_ENV === "development";
  const lines = parseStack(error.stack);
  const appLines = lines.filter((l) => l.isApp);

  useEffect(() => {
    logError(error, "Dashboard");
  }, [error]);

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
          {/* Component Stack — most useful, shown first */}
          {error.componentStack && (
            <div className="rounded-lg border bg-muted/30 overflow-hidden">
              <button
                className="w-full flex items-center justify-between px-4 py-2.5 text-xs font-medium text-muted-foreground hover:text-foreground transition-colors"
                onClick={() => setCompExpanded((v) => !v)}
              >
                <span>
                  📍 Component stack
                  <span className="ml-2 text-emerald-500 font-semibold">← exact component tree</span>
                </span>
                {compExpanded ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
              </button>
              {compExpanded && (
                <pre className="bg-slate-950 px-4 py-3 text-emerald-400 text-xs overflow-x-auto max-h-60 overflow-y-auto whitespace-pre-wrap">
                  {error.componentStack.trim()}
                </pre>
              )}
            </div>
          )}

          {/* JS Stack trace */}
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

              {/* Always show first app frame as quick hint */}
              {!stackExpanded && appLines.length > 0 && (
                <div className="bg-slate-950 px-4 py-2 text-orange-400 font-semibold text-xs border-t border-slate-800">
                  {appLines[0].raw}
                </div>
              )}

              {stackExpanded && (
                <div className="bg-slate-950 px-4 py-3 overflow-x-auto max-h-72 overflow-y-auto">
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
