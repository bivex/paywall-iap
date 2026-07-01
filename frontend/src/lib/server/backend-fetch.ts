import { cookies } from "next/headers";
import { type NextRequest } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export async function backendFetch(
  path: string,
  req: NextRequest | null,
  init?: RequestInit
): Promise<Response> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) throw new Error("UNAUTHORIZED");

  const appId = req?.nextUrl?.searchParams?.get("app_id") ?? null;

  const headers: Record<string, string> = {
    ...((init?.headers ?? {}) as Record<string, string>),
    Authorization: `Bearer ${token}`,
  };
  if (appId) headers["X-App-ID"] = appId;

  return fetch(`${BACKEND_URL}${path}`, { ...init, headers, cache: "no-store" });
}
