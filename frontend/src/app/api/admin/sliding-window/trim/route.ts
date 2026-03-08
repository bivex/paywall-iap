import { NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as { experimentId?: string } | null;

  if (!body?.experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  const res = await fetch(`${BACKEND_URL}/v1/bandit/experiments/${body.experimentId}/window/trim`, {
    method: "POST",
  });

  const responseBody = await res.json().catch(() => ({}));
  return NextResponse.json(responseBody, { status: res.status });
}
