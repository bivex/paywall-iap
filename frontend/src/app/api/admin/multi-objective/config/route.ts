import { NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export async function PUT(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    experimentId?: string;
    objectiveType?: string;
    objectiveWeights?: Record<string, number>;
  } | null;

  if (!body?.experimentId || !body?.objectiveType) {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 });
  }

  const res = await fetch(`${BACKEND_URL}/v1/bandit/experiments/${body.experimentId}/objectives/config`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      objective_type: body.objectiveType,
      objective_weights: body.objectiveWeights,
    }),
  });

  const responseBody = await res.json().catch(() => ({}));
  return NextResponse.json(responseBody, { status: res.status });
}
