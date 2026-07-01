import { NextRequest, NextResponse } from "next/server";

import { getMultiObjectiveDashboardFromCookies } from "@/lib/server/multi-objective-admin";

export async function GET(req: NextRequest) {
  const appId = req.nextUrl.searchParams.get("app_id");
  const data = await getMultiObjectiveDashboardFromCookies(appId);
  return NextResponse.json({ data });
}
