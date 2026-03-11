import { NextRequest, NextResponse } from "next/server";

const studentRoutes = ["/island", "/game"];
const dashboardRoutes = ["/overview"];
const setupRoute = "/setup";

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  const isStudentRoute = studentRoutes.some((r) => pathname.startsWith(r));
  const isDashboardRoute = dashboardRoutes.some((r) => pathname.startsWith(r));
  const isSetupRoute = pathname.startsWith(setupRoute);

  // Public routes — no protection needed
  if (!isStudentRoute && !isDashboardRoute && !isSetupRoute) {
    return NextResponse.next();
  }

  // For protected routes, we redirect to login if no auth data exists.
  // Since tokens are stored in localStorage (client-side), we use a special header
  // that client-side components set, or we let the client handle the redirect.
  // 
  // The middleware will add a response header indicating this is a protected route,
  // and client-side auth components will handle the actual auth check.
  
  const response = NextResponse.next();
  response.headers.set("X-Protected-Route", "true");
  return response;
}

export const config = {
  matcher: ["/island/:path*", "/game/:path*", "/overview/:path*", "/setup/:path*"],
};
