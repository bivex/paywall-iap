import { type NextRequest, NextResponse } from "next/server";

const PUBLIC_PATHS = ["/auth/", "/api/"];

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));
  const token = req.cookies.get("admin_access_token")?.value;

  // Unauthenticated user trying to access protected route
  if (!isPublic && !token) {
    const loginUrl = req.nextUrl.clone();
    loginUrl.pathname = "/auth/v1/login";
    loginUrl.search = "";
    return NextResponse.redirect(loginUrl);
  }

  // Authenticated user trying to access login page
  if (pathname.startsWith("/auth/") && token) {
    const dashboardUrl = req.nextUrl.clone();
    dashboardUrl.pathname = "/dashboard/default";
    dashboardUrl.search = "";
    return NextResponse.redirect(dashboardUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|ico|webp)$).*)"],
};
