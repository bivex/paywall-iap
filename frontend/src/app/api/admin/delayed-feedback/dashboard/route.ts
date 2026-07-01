import { NextRequest, NextResponse } from "next/server";

import { getDelayedFeedbackDashboardFromCookies } from "@/lib/server/delayed-feedback-admin";

export async function GET(req: NextRequest) {
  const appId = req.nextUrl.searchParams.get("app_id");
  const data = await getDelayedFeedbackDashboardFromCookies(appId);
  return NextResponse.json({ data });
}
