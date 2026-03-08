"use client";

import { useEffect } from "react";

import { normalizeUnknownError, reportJsError } from "@/lib/js-error-reporter.client";

export function JsErrorCatcher() {
  useEffect(() => {
    const onError = (event: ErrorEvent) => {
      const normalized = normalizeUnknownError(event.error ?? event.message);
      reportJsError({
        type: "error",
        name: normalized.name,
        message: normalized.message,
        stack: normalized.stack,
        source: event.filename,
        lineno: event.lineno,
        colno: event.colno,
      });
    };

    const onUnhandledRejection = (event: PromiseRejectionEvent) => {
      const normalized = normalizeUnknownError(event.reason);
      reportJsError({
        type: "unhandledrejection",
        name: normalized.name,
        message: normalized.message,
        stack: normalized.stack,
      });
    };

    window.addEventListener("error", onError);
    window.addEventListener("unhandledrejection", onUnhandledRejection);

    return () => {
      window.removeEventListener("error", onError);
      window.removeEventListener("unhandledrejection", onUnhandledRejection);
    };
  }, []);

  return null;
}
