"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export interface UserRow {
  id: string;
  platform_user_id: string;
  platform: string;
  email: string;
  role: string;
  ltv: number;
  app_version: string;
  created_at: string;
  sub_status: string;
  sub_expires_at: string;
}

export interface UsersResponse {
  users: UserRow[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

const EMPTY: UsersResponse = { users: [], total: 0, page: 1, limit: 20, total_pages: 0 };

export async function getUsers(params: {
  page?: number;
  limit?: number;
  search?: string;
  platform?: string;
  role?: string;
} = {}): Promise<UsersResponse> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return EMPTY;

  const qs = new URLSearchParams();
  if (params.page)     qs.set("page", String(params.page));
  if (params.limit)    qs.set("limit", String(params.limit));
  if (params.search)   qs.set("search", params.search);
  if (params.platform) qs.set("platform", params.platform);
  if (params.role)     qs.set("role", params.role);

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/users/search?${qs}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return EMPTY;
    return (await res.json()) as UsersResponse;
  } catch {
    return EMPTY;
  }
}
