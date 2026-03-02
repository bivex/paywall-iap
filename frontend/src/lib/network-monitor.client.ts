/**
 * Dev-only fetch interceptor.
 * Patches globalThis.fetch to record failed requests (non-2xx or thrown).
 * Call installNetworkMonitor() once in a client layout/component.
 * Call getFailedRequests() in error.tsx to read captured failures.
 */

export interface FailedRequest {
  method: string;
  url: string;
  status: number;
  body: string;
  time: string;
}

const MAX = 10;
const _failed: FailedRequest[] = [];
let _installed = false;

export function installNetworkMonitor(): void {
  if (typeof window === "undefined") return;
  if (process.env.NODE_ENV !== "development") return;
  if (_installed) return;
  _installed = true;

  const _origFetch = window.fetch.bind(window);

  window.fetch = async (input, init) => {
    const url =
      typeof input === "string"
        ? input
        : input instanceof URL
          ? input.toString()
          : (input as Request).url;
    const method = (init?.method ?? (input instanceof Request ? input.method : "GET")).toUpperCase();

    try {
      const res = await _origFetch(input, init);

      if (!res.ok) {
        let body = "";
        try {
          // Clone so the caller can still read the body
          const clone = res.clone();
          const text = await clone.text();
          body = text.length > 200 ? text.slice(0, 200) + "…" : text;
        } catch {
          // ignore
        }
        record({ method, url, status: res.status, body });
      }

      return res;
    } catch (err) {
      record({ method, url, status: 0, body: String(err) });
      throw err;
    }
  };
}

function record(entry: Omit<FailedRequest, "time">) {
  const full: FailedRequest = { ...entry, time: new Date().toLocaleTimeString() };
  _failed.unshift(full);
  if (_failed.length > MAX) _failed.pop();
}

/** Returns a snapshot of recently failed requests (newest first). */
export function getFailedRequests(): FailedRequest[] {
  return [..._failed];
}

/** Clear recorded failures (e.g. after reset). */
export function clearFailedRequests(): void {
  _failed.length = 0;
}
