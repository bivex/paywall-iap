import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/server/backend-fetch";

export async function GET(req: NextRequest) {
  const { searchParams } = req.nextUrl;
  const experimentId = searchParams.get("experimentId");
  const limit = searchParams.get("limit");

  if (!experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  try {
    const backendPath = `/v1/bandit/experiments/${experimentId}/window/events${limit ? `?limit=${limit}` : ""}`;
    const res = await backendFetch(backendPath, req);
    const body = await res.json().catch(() => ({}));
    return NextResponse.json(body, { status: res.status });
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
}
