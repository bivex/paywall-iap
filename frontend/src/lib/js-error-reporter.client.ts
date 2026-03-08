"use client";

type ReportPayload = {
  type: "error" | "unhandledrejection" | "global-error";
  message: string;
  name?: string;
  stack?: string;
  componentStack?: string;
  digest?: string;
  source?: string;
  lineno?: number;
  colno?: number;
  href: string;
  userAgent: string;
  time: string;
};

const COLLECTOR_URL =
  process.env.NEXT_PUBLIC_JS_ERROR_COLLECTOR_URL ?? "http://localhost:8088/frontend-error";

const recentFingerprints = new Map<string, number>();
const DEDUPE_WINDOW_MS = 5_000;

export function reportJsError(payload: Omit<ReportPayload, "href" | "userAgent" | "time">): void {
  if (typeof window === "undefined") return;

  const fullPayload: ReportPayload = {
    ...payload,
    href: window.location.href,
    userAgent: navigator.userAgent,
    time: new Date().toISOString(),
  };

  const fingerprint = `${fullPayload.type}:${fullPayload.message}:${fullPayload.stack ?? ""}:${fullPayload.source ?? ""}`;
  const now = Date.now();
  const prev = recentFingerprints.get(fingerprint) ?? 0;
  if (now - prev < DEDUPE_WINDOW_MS) return;
  recentFingerprints.set(fingerprint, now);

  const body = JSON.stringify(fullPayload);

  try {
    if (typeof navigator.sendBeacon === "function") {
      const blob = new Blob([body], { type: "text/plain;charset=UTF-8" });
      if (navigator.sendBeacon(COLLECTOR_URL, blob)) return;
    }
  } catch {
    // fall through to fetch
  }

  void fetch(COLLECTOR_URL, {
    method: "POST",
    mode: "no-cors",
    keepalive: true,
    headers: { "Content-Type": "text/plain;charset=UTF-8" },
    body,
  }).catch(() => {
    // ignore reporting failures
  });
}

export function normalizeUnknownError(error: unknown): { name?: string; message: string; stack?: string } {
  if (error instanceof Error) {
    return {
      name: error.name,
      message: error.message,
      stack: error.stack,
    };
  }

  return {
    message: typeof error === "string" ? error : JSON.stringify(error),
  };
}