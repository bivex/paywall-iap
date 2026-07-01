import { NextRequest, NextResponse } from "next/server";

import { backendFetch } from "@/lib/server/backend-fetch";

export async function POST(req: NextRequest) {
  const body = (await req.json().catch(() => null)) as {
    transactionId?: string;
    userId?: string;
    conversionValue?: number;
    currency?: string;
  } | null;

  if (!body?.transactionId || !body?.userId || typeof body.conversionValue !== "number" || !body?.currency) {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 });
  }

  try {
    const res = await backendFetch("/v1/bandit/conversions", req, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        transaction_id: body.transactionId,
        user_id: body.userId,
        conversion_value: body.conversionValue,
        currency: body.currency,
      }),
    });
    const responseBody = await res.json().catch(() => ({}));
    return NextResponse.json(responseBody, { status: res.status });
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }
}
