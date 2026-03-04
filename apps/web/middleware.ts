import { NextRequest, NextResponse } from "next/server";

const studentRoutes = ["/island", "/game"];
const dashboardRoutes = ["/overview"];
const setupRoute = "/setup";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const accessToken = request.cookies.get("access_token")?.value;

  const isStudentRoute = studentRoutes.some((r) => pathname.startsWith(r));
  const isDashboardRoute = dashboardRoutes.some((r) => pathname.startsWith(r));
  const isSetupRoute = pathname.startsWith(setupRoute);

  // Public routes — no protection needed
  if (!isStudentRoute && !isDashboardRoute && !isSetupRoute) {
    return NextResponse.next();
  }

  // No token → redirect to login
  if (!accessToken) {
    const loginUrl = new URL("/login", request.url);
    return NextResponse.redirect(loginUrl);
  }

  // Decode JWT payload for UX-only routing. Signature verification happens
  // server-side on every API call (internal/auth middleware). This decode is
  // solely to redirect users to the correct section — it is NOT a security
  // gate. If the token is tampered with, server APIs will reject it.
  let role: string | undefined;
  try {
    const payloadBase64 = accessToken.split(".")[1];
    if (!payloadBase64) {
      return NextResponse.redirect(new URL("/login", request.url));
    }
    const payload = JSON.parse(
      Buffer.from(payloadBase64, "base64").toString("utf-8")
    );
    role = payload.role;

    if (!role) {
      return NextResponse.redirect(new URL("/login", request.url));
    }

    // Students cannot access dashboard routes
    if (isDashboardRoute && role === "student") {
      return NextResponse.redirect(new URL("/island", request.url));
    }

    // Parents and educators cannot access student routes or setup
    if ((isStudentRoute || isSetupRoute) && (role === "parent" || role === "educator")) {
      return NextResponse.redirect(new URL("/overview", request.url));
    }
  } catch {
    // Malformed token → redirect to login
    const loginUrl = new URL("/login", request.url);
    return NextResponse.redirect(loginUrl);
  }

  // For student-role users, check profile existence to decide setup vs island.
  // This is a UX convenience — the pages themselves also verify.
  if (role === "student" && (isStudentRoute || isSetupRoute)) {
    try {
      const profileRes = await fetch(`${API_URL}/api/profiles/me`, {
        headers: { Cookie: `access_token=${accessToken}` },
      });

      if (profileRes.status === 404 && !isSetupRoute) {
        // No profile yet → redirect to setup
        return NextResponse.redirect(new URL("/setup", request.url));
      }

      if (profileRes.ok && isSetupRoute) {
        // Already has profile → redirect to island
        return NextResponse.redirect(new URL("/island", request.url));
      }
    } catch {
      // API unreachable — allow through; page-level checks will handle it
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/island/:path*", "/game/:path*", "/overview/:path*", "/setup/:path*"],
};
