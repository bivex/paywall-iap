import { NextResponse } from "next/server";

import { getBanditDashboardFromCookies } from "@/lib/server/bandit-admin";

export async function GET() {
  const data = await getBanditDashboardFromCookies();
  return NextResponse.json({ data });
}
