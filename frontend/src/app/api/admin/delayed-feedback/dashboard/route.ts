import { NextResponse } from "next/server";

import { getDelayedFeedbackDashboardFromCookies } from "@/lib/server/delayed-feedback-admin";

export async function GET() {
  const data = await getDelayedFeedbackDashboardFromCookies();
  return NextResponse.json({ data });
}
