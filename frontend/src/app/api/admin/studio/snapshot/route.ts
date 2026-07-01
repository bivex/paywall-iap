import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { getStudioSnapshotFromCookies } from "@/lib/server/studio-admin";

export async function GET(req: NextRequest) {
  const { searchParams } = req.nextUrl;
  const experimentId = searchParams.get("experimentId");
  const appId =
    searchParams.get("app_id") ??
    (await cookies()).get("admin_app_id")?.value ??
    null;

  if (!experimentId) {
    return NextResponse.json({ error: "experimentId is required" }, { status: 400 });
  }

  const data = await getStudioSnapshotFromCookies(experimentId, appId);
  if (!data) {
    return NextResponse.json({ error: "Unable to load studio snapshot" }, { status: 404 });
  }

  return NextResponse.json({ data });
}
