"use server";

import { isFetchError, serverFetch } from "@/lib/server-fetch";

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
  plan_type: string;
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

export async function getUsers(
  params: { page?: number; limit?: number; search?: string; platform?: string; role?: string } = {},
): Promise<UsersResponse> {
  const qs = new URLSearchParams();
  if (params.page) qs.set("page", String(params.page));
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.search) qs.set("search", params.search);
  if (params.platform) qs.set("platform", params.platform);
  if (params.role) qs.set("role", params.role);

  const res = await serverFetch<UsersResponse>(`/v1/admin/users/search?${qs}`);
  if (isFetchError(res)) {
    return EMPTY;
  }
  return res;
}
