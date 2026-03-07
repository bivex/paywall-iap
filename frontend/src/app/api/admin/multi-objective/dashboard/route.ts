import { NextResponse } from "next/server";

import { getMultiObjectiveDashboardFromCookies } from "@/lib/server/multi-objective-admin";

export async function GET() {
  const data = await getMultiObjectiveDashboardFromCookies();
  return NextResponse.json({ data });
}
