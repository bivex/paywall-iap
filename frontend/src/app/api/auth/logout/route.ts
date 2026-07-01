import { cookies } from "next/headers";
import { NextResponse } from "next/server";

/**
 * Clears the admin session cookies and redirects to the login page.
 *
 * Used as the redirect target when a Server Component detects an expired /
 * invalid token (HTTP 401). Cookies cannot be deleted during RSC render
 * ("Cookies can only be modified in a Server Action or Route Handler"), so the
 * server redirects here instead, where mutation is allowed.
 *
 * Clearing the cookie is required: src/proxy.ts bounces any request that still
 * carries a token away from /auth/*, so leaving a stale token in place would
 * cause a redirect loop back into the dashboard.
 *
 * Reaches via GET so a plain `redirect("/api/auth/logout")` from the server works.
 */
export async function GET(req: Request) {
  const c = await cookies();
  c.delete("admin_access_token");
  c.delete("admin_refresh_token");
  c.delete("admin_email");
  c.delete("admin_role");
  return NextResponse.redirect(new URL("/auth/v1/login", req.url), { status: 302 });
}
