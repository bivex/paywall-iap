import { NextResponse } from "next/server";

import { getStudioDashboardFromCookies } from "@/lib/server/studio-admin";

export async function GET() {
  const data = await getStudioDashboardFromCookies();
  return NextResponse.json({ data });
}
