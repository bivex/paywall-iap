import { NextRequest, NextResponse } from "next/server";

import { getStudioDashboardFromCookies } from "@/lib/server/studio-admin";

export async function GET(req: NextRequest) {
  const appId = req.nextUrl.searchParams.get("app_id");
  const data = await getStudioDashboardFromCookies(appId);
  return NextResponse.json({ data });
}
