import { NextRequest, NextResponse } from "next/server";

const studentRoutes = ["/island", "/game"];
const dashboardRoutes = ["/overview"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const accessToken = request.cookies.get("access_token")?.value;

  const isStudentRoute = studentRoutes.some((r) => pathname.startsWith(r));
  const isDashboardRoute = dashboardRoutes.some((r) => pathname.startsWith(r));

  // Public routes — no protection needed
  if (!isStudentRoute && !isDashboardRoute) {
    return NextResponse.next();
  }

  // No token → redirect to login
  if (!accessToken) {
    const loginUrl = new URL("/login", request.url);
    return NextResponse.redirect(loginUrl);
  }

  // Decode JWT payload to check role (without verifying — verification is server-side)
  try {
    const payloadBase64 = accessToken.split(".")[1];
    const payload = JSON.parse(
      Buffer.from(payloadBase64, "base64").toString("utf-8")
    );
    const role: string = payload.role;

    // Students cannot access dashboard routes
    if (isDashboardRoute && role === "student") {
      return NextResponse.redirect(new URL("/island", request.url));
    }

    // Parents and educators cannot access student routes
    if (isStudentRoute && (role === "parent" || role === "educator")) {
      return NextResponse.redirect(new URL("/overview", request.url));
    }
  } catch {
    // Malformed token → redirect to login
    const loginUrl = new URL("/login", request.url);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/island/:path*", "/game/:path*", "/overview/:path*"],
};
