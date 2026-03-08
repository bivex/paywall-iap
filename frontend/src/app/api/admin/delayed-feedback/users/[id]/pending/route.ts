import { NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://api:8080";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params;
  if (!id) {
    return NextResponse.json({ error: "id is required" }, { status: 400 });
  }

  const res = await fetch(`${BACKEND_URL}/v1/bandit/users/${id}/pending`, {
    cache: "no-store",
  });

  const responseBody = await res.json().catch(() => ({}));
  return NextResponse.json(responseBody, { status: res.status });
}