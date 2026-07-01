import { NextRequest, NextResponse } from "next/server";

import { getSlidingWindowDashboardFromCookies } from "@/lib/server/sliding-window-admin";

export async function GET(req: NextRequest) {
  const appId = req.nextUrl.searchParams.get("app_id");
  const data = await getSlidingWindowDashboardFromCookies(appId);
  return NextResponse.json({ data });
}
