import { NextRequest, NextResponse } from "next/server";

import { getBanditDashboardFromCookies } from "@/lib/server/bandit-admin";

export async function GET(req: NextRequest) {
  const appId = req.nextUrl.searchParams.get("app_id");
  const data = await getBanditDashboardFromCookies(appId);
  return NextResponse.json({ data });
}
