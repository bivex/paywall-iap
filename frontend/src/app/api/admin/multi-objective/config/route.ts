import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/server/backend-fetch";

export async function PUT(req: NextRequest) {
  const body = (await req.json().catch(() => null)) as {
    experimentId?: string;
    objectiveType?: string;
    objectiveWeights?: Record<string, number>;
  } | null;

  if (!body?.experimentId || !body?.objectiveType) {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 });
  }

  try {
    const res = await backendFetch(
      `/v1/bandit/experiments/${body.experimentId}/objectives/config`,
      req,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          objective_type: body.objectiveType,
          objective_weights: body.objectiveWeights,
        }),
      }
    );
    const responseBody = await res.json().catch(() => ({}));
    return NextResponse.json(responseBody, { status: res.status });
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
}
