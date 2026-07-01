import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { getMultiObjectiveDashboardFromCookies } from "@/lib/server/multi-objective-admin";

export async function GET(req: NextRequest) {
  const appId =
    req.nextUrl.searchParams.get("app_id") ??
    (await cookies()).get("admin_app_id")?.value ??
    null;
  const data = await getMultiObjectiveDashboardFromCookies(appId);
  return NextResponse.json({ data });
}
