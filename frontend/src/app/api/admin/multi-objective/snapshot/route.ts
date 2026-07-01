import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { getMultiObjectiveSnapshotFromCookies } from "@/lib/server/multi-objective-admin";

export async function GET(req: NextRequest) {
  const { searchParams } = req.nextUrl;
  const experimentId = searchParams.get("experimentId");
  const appId = searchParams.get("app_id");

  if (!experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  const data = await getMultiObjectiveSnapshotFromCookies(experimentId, appId);
  if (!data) {
    return NextResponse.json({ error: "Unable to load multi-objective snapshot" }, { status: 404 });
  }

  return NextResponse.json({ data });
}
