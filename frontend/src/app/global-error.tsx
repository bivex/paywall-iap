"use client";

import { useEffect } from "react";

import { reportJsError } from "@/lib/js-error-reporter.client";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string; componentStack?: string };
  reset: () => void;
}) {
  return (
    <html lang="en">
      <body>
        <ErrorDisplay error={error} reset={reset} />
      </body>
    </html>
  );
}

function ErrorDisplay({
  error,
  reset,
}: {
  error: Error & { digest?: string; componentStack?: string };
  reset: () => void;
}) {
  useEffect(() => {
    reportJsError({
      type: "global-error",
      name: error.name,
      message: error.message,
      stack: error.stack,
      componentStack: error.componentStack,
      digest: error.digest,
    });
  }, [error]);

  const isDev = process.env.NODE_ENV === "development";
  const lines = parseStack(error.stack);

  return (
    <div style={{ fontFamily: "monospace", padding: 32, maxWidth: 900, margin: "0 auto" }}>
      <div style={{ marginBottom: 16, display: "flex", alignItems: "center", gap: 12 }}>
        <span style={{ fontSize: 24 }}>💥</span>
        <strong style={{ fontSize: 18, color: "#dc2626" }}>Unhandled Error</strong>
        {error.digest && (
          <code style={{ fontSize: 11, color: "#6b7280", background: "#f3f4f6", padding: "2px 6px", borderRadius: 4 }}>
            digest: {error.digest}
          </code>
        )}
      </div>

      <div
        style={{ background: "#fef2f2", border: "1px solid #fca5a5", borderRadius: 8, padding: 16, marginBottom: 16 }}
      >
        <div style={{ color: "#991b1b", fontWeight: 600, marginBottom: 4 }}>{error.name}</div>
        <div style={{ color: "#dc2626", fontSize: 14 }}>{error.message}</div>
      </div>

      {isDev && error.componentStack && (
        <details open style={{ marginBottom: 16 }}>
          <summary style={{ cursor: "pointer", color: "#059669", fontWeight: 600, fontSize: 12, marginBottom: 8 }}>
            📍 Component stack (most precise)
          </summary>
          <pre
            style={{
              background: "#0f172a",
              color: "#34d399",
              borderRadius: 8,
              padding: 16,
              fontSize: 11,
              overflow: "auto",
              maxHeight: 240,
              whiteSpace: "pre-wrap",
            }}
          >
            {error.componentStack.trim()}
          </pre>
        </details>
      )}

      {isDev && lines.length > 0 && (
        <details style={{ marginBottom: 16 }}>
          <summary style={{ cursor: "pointer", color: "#94a3b8", fontSize: 12, marginBottom: 8 }}>
            JS stack trace
          </summary>
          <div style={{ background: "#0f172a", borderRadius: 8, padding: 16, overflowX: "auto" }}>
            {lines.map((line) => (
              <div key={line.raw} style={{ marginBottom: 2 }}>
                {line.isApp ? (
                  <span style={{ color: "#f97316", fontWeight: 600, fontSize: 12 }}>{line.raw}</span>
                ) : (
                  <span style={{ color: "#64748b", fontSize: 11 }}>{line.raw}</span>
                )}
              </div>
            ))}
          </div>
        </details>
      )}

      <button
        type="button"
        onClick={reset}
        style={{
          background: "#2563eb",
          color: "#fff",
          border: "none",
          borderRadius: 6,
          padding: "8px 20px",
          fontSize: 14,
          cursor: "pointer",
        }}
      >
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
      isApp: raw.includes("src/") || raw.includes("app/") || raw.includes("actions/"),
    }));
}
