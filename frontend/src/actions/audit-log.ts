"use server";

import { cookies } from "next/headers";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

export interface AuditLogRow {
  ID: string;
  Time: string;
  AdminEmail: string;
  Action: string;
  TargetType: string;
  Detail: string;
  IPAddress: string;
}

export interface AuditLogResponse {
  rows: AuditLogRow[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface AuditLogParams {
  page?: number;
  limit?: number;
  action?: string;
  search?: string;
  from?: string;
  to?: string;
}

const EMPTY: AuditLogResponse = {
  rows: [],
  total: 0,
  page: 1,
  limit: 20,
  total_pages: 0,
};

export async function getAuditLog(params: AuditLogParams = {}): Promise<AuditLogResponse> {
  const cookieStore = await cookies();
  const token = cookieStore.get("admin_access_token")?.value;
  if (!token) return EMPTY;

  const qs = new URLSearchParams();
  if (params.page) qs.set("page", String(params.page));
  if (params.limit) qs.set("limit", String(params.limit));
  if (params.action) qs.set("action", params.action);
  if (params.search) qs.set("search", params.search);
  if (params.from) qs.set("from", params.from);
  if (params.to) qs.set("to", params.to);

  try {
    const res = await fetch(`${BACKEND_URL}/v1/admin/audit-log?${qs.toString()}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    });
    if (!res.ok) return EMPTY;
    return (await res.json()) as AuditLogResponse;
  } catch {
    return EMPTY;
  }
}
