import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/server/backend-fetch";

export async function POST(req: NextRequest) {
  const body = (await req.json().catch(() => null)) as { experimentId?: string } | null;

  if (!body?.experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  try {
    const res = await backendFetch(
      `/v1/bandit/experiments/${body.experimentId}/window/trim`,
      req,
      { method: "POST" }
    );
    const responseBody = await res.json().catch(() => ({}));
    return NextResponse.json(responseBody, { status: res.status });
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
}
