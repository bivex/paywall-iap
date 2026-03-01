"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export async function loginAction(email: string, password: string): Promise<{ error?: string; redirectTo?: string }> {
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
  const data = (body.data ?? body) as { access_token: string; refresh_token?: string };

  const cookieStore = await cookies();
  cookieStore.set("admin_access_token", data.access_token, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 15, // 15 min (matches backend access token TTL)
  });

  if (data.refresh_token) {
    cookieStore.set("admin_refresh_token", data.refresh_token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: 60 * 60 * 24 * 30, // 30 days
    });
  }

  return { redirectTo: "/dashboard/default" };
}

export async function logoutAction(): Promise<void> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;

  if (token) {
    await fetch(`${BACKEND_URL}/v1/admin/auth/logout`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
    }).catch(() => {});
  }

  cookieStore.delete("admin_access_token");
  cookieStore.delete("admin_refresh_token");

  redirect("/auth/v1/login");
}
