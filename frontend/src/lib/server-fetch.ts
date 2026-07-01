import { cookies } from "next/headers";
import { redirect } from "next/navigation";

/** Where unauthenticated users are sent (matches src/proxy.ts). */
export const LOGIN_PATH = "/auth/v1/login";

/**
 * Route handler that clears session cookies then redirects to LOGIN_PATH.
 * Used on 401 because cookies can't be deleted during RSC render.
 */
export const LOGOUT_PATH = "/api/auth/logout";

/** A recoverable load failure (5xx, network, …) — NOT an auth problem. */
export interface FetchError {
  error: true;
  message: string;
  status: number;
}

/** Either the parsed payload T, or a recoverable FetchError. */
export type ServerFetchResult<T> = T | FetchError;

/** Type guard: true when a serverFetch call returned a recoverable error. */
export function isFetchError<T>(result: ServerFetchResult<T>): result is FetchError {
  return typeof result === "object" && result !== null && (result as FetchError).error === true;
}

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

/**
 * Read the admin session token and selected app id from cookies, for Server
 * Actions that call the multi-tenant backend directly (so they can attach the
 * X-App-ID header without going through serverFetch). `token`/`appId` are
 * undefined when absent.
 */
export async function getAuth(): Promise<{ token?: string; appId?: string }> {
  const cookieStore = await cookies();
  return {
    token: cookieStore.get("admin_access_token")?.value,
    appId: cookieStore.get("admin_app_id")?.value,
  };
}

/**
 * Authenticated server-side fetch for admin API calls.
 *
 * Auth handling (callers never check the token themselves):
 *  - no token   → redirect to the login page
 *  - HTTP 401   → clear the stale cookie and redirect to the login page
 *
 * Recoverable errors are returned (not thrown) as FetchError so the page can
 * render a "Failed to load / Retry" state for an authenticated user:
 *  - network failure  → { status: 0, … }
 *  - other non-2xx    → { status: <code>, … }
 *
 * Success → parsed JSON typed as T.
 *
 * Use only inside Server Components / Server Actions (reads cookies, redirects).
 */
export async function serverFetch<T>(path: string, init?: RequestInit): Promise<ServerFetchResult<T>> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) {
    redirect(LOGIN_PATH);
  }
  const appId = cookieStore.get("admin_app_id")?.value;

  let res: Response;
  try {
    res = await fetch(`${BACKEND_URL}${path}`, {
      cache: "no-store",
      ...init,
      headers: {
        Authorization: `Bearer ${token}`,
        ...(appId ? { "X-App-ID": appId } : {}),
        ...(init?.headers ?? {}),
      },
    });
  } catch {
    return { error: true, message: "Could not reach the server. Please try again.", status: 0 };
  }

  if (res.status === 401) {
    // Session expired / invalid token. Cookies can't be deleted during RSC
    // render, so hand off to the logout route handler, which clears them and
    // redirects to login (also breaks proxy's "token-present → leave /auth" loop).
    redirect(LOGOUT_PATH);
  }

  if (!res.ok) {
    return { error: true, message: `Failed to load data (HTTP ${res.status}). Please try again.`, status: res.status };
  }

  return (await res.json()) as T;
}
