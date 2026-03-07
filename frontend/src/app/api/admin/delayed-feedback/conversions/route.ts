import { NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export async function POST(request: Request) {
  const body = (await request.json().catch(() => null)) as {
    transactionId?: string;
    userId?: string;
    conversionValue?: number;
    currency?: string;
  } | null;

  if (!body?.transactionId || !body?.userId || typeof body.conversionValue !== "number" || !body?.currency) {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 });
  }

  const res = await fetch(`${BACKEND_URL}/v1/bandit/conversions`, {
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
}
