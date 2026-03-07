import { NextResponse } from "next/server";

import { getStudioSnapshotFromCookies } from "@/lib/server/studio-admin";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const experimentId = searchParams.get("experimentId");

  if (!experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  const data = await getStudioSnapshotFromCookies(experimentId);
  if (!data) {
    return NextResponse.json({ error: "Unable to load studio snapshot" }, { status: 404 });
  }

  return NextResponse.json({ data });
}
