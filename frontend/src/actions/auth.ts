/**
 * Copyright (c) 2026 Bivex
 *
 * Author: Bivex
 * Available for contact via email: support@b-b.top
 * For up-to-date contact information:
 * https://github.com/bivex
 *
 * Created: 2026-03-07 22:20
 * Last Updated: 2026-03-07 22:20
 *
 * Licensed under the MIT License.
 * Commercial licensing available upon request.
 */

"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8081";
// Cookies are secure only when served over HTTPS (set HTTPS=true in production with TLS)
const SECURE_COOKIES = process.env.HTTPS === "true";

export async function loginAction(
  email: string,
  password: string,
  remember = false,
): Promise<{ error?: string; redirectTo?: string }> {
  let res: Response;
  try {
    res = await fetch(`${BACKEND_URL}/v1/admin/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
  } catch {
    return { error: "Cannot connect to server. Check your network." };
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return { error: (body as { error?: string }).error ?? "Invalid credentials." };
  }

  const body = await res.json();
  const data = (body.data ?? body) as { access_token: string; refresh_token?: string; email?: string; role?: string };

  const THIRTY_DAYS = 60 * 60 * 24 * 30;
  const cookieStore = await cookies();
  cookieStore.set("admin_access_token", data.access_token, {
    httpOnly: true,
    secure: SECURE_COOKIES,
    sameSite: "lax",
    path: "/",
    // remember=true → persist 30 days; false → session cookie (expires on browser close)
    ...(remember ? { maxAge: THIRTY_DAYS } : {}),
  });

  const bodyData = data;
  if (bodyData.email) {
    cookieStore.set("admin_email", bodyData.email, {
      httpOnly: false,
      secure: SECURE_COOKIES,
      sameSite: "lax",
      path: "/",
      ...(remember ? { maxAge: THIRTY_DAYS } : {}),
    });
  }
  if (bodyData.role) {
    cookieStore.set("admin_role", bodyData.role, {
      httpOnly: false,
      secure: SECURE_COOKIES,
      sameSite: "lax",
      path: "/",
      ...(remember ? { maxAge: THIRTY_DAYS } : {}),
    });
  }

  if (data.refresh_token) {
    cookieStore.set("admin_refresh_token", data.refresh_token, {
      httpOnly: true,
      secure: SECURE_COOKIES,
      sameSite: "lax",
      path: "/",
      ...(remember ? { maxAge: THIRTY_DAYS } : {}),
    });
  }

  return { redirectTo: "/dashboard/default" };
}

export async function logoutAction(): Promise<{ redirectTo: string }> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;

  if (token) {
    await fetch(`${BACKEND_URL}/v1/admin/auth/logout`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    }).catch(() => undefined);
  }

  cookieStore.delete("admin_access_token");
  cookieStore.delete("admin_refresh_token");
  cookieStore.delete("admin_email");
  cookieStore.delete("admin_role");

  return { redirectTo: "/auth/v1/login" };
}
