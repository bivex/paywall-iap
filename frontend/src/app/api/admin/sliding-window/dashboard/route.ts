import { NextResponse } from "next/server";

import { getSlidingWindowDashboardFromCookies } from "@/lib/server/sliding-window-admin";

export async function GET() {
  const data = await getSlidingWindowDashboardFromCookies();
  return NextResponse.json({ data });
}
