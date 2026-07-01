import { NextRequest, NextResponse } from "next/server";

import { getDelayedFeedbackSnapshotFromCookies } from "@/lib/server/delayed-feedback-admin";

export async function GET(req: NextRequest) {
  const { searchParams } = req.nextUrl;
  const experimentId = searchParams.get("experimentId");
  const appId = searchParams.get("app_id");

  if (!experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  const data = await getDelayedFeedbackSnapshotFromCookies(experimentId, appId);
  if (!data) {
    return NextResponse.json({ error: "Unable to load delayed feedback snapshot" }, { status: 404 });
  }

  return NextResponse.json({ data });
}
