import { NextResponse } from "next/server";

import { getSlidingWindowSnapshotFromCookies } from "@/lib/server/sliding-window-admin";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const experimentId = searchParams.get("experimentId");

  if (!experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  const data = await getSlidingWindowSnapshotFromCookies(experimentId);
  if (!data) {
    return NextResponse.json({ error: "Unable to load sliding window snapshot" }, { status: 404 });
  }

  return NextResponse.json({ data });
}
