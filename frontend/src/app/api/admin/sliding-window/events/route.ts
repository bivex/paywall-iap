import { NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const experimentId = searchParams.get("experimentId");
  const limit = searchParams.get("limit");

  if (!experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  const backendUrl = new URL(`${BACKEND_URL}/v1/bandit/experiments/${experimentId}/window/events`);
  if (limit) {
    backendUrl.searchParams.set("limit", limit);
  }

  const res = await fetch(backendUrl, { cache: "no-store" });
  const responseBody = await res.json().catch(() => ({}));
  return NextResponse.json(responseBody, { status: res.status });
}
