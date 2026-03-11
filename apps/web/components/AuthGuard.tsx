"use client";

import { useEffect, useState } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuthStore } from "@/hooks/useAuthStore";
import { setAuthCallbacks } from "@/lib/api";

interface AuthGuardProps {
  children: React.ReactNode;
  allowedRoles?: ("student" | "parent" | "educator")[];
}

export function AuthGuard({ children, allowedRoles }: AuthGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { accessToken, user, fetchUser } = useAuthStore();
  const [isInitialized, setIsInitialized] = useState(false);

  // Initialize auth callbacks once
  useEffect(() => {
    setAuthCallbacks(
      () => useAuthStore.getState().accessToken,
      () => useAuthStore.getState().refreshTokens()
    );
    setIsInitialized(true);
  }, []);

  useEffect(() => {
    if (!isInitialized) return;

    // No access token - redirect to login
    if (!accessToken) {
      router.push("/login");
      return;
    }

    // If we have token but no user, fetch user data
    if (accessToken && !user) {
      fetchUser().catch(() => {
        router.push("/login");
      });
      return;
    }

    // Check role-based access
    if (user && allowedRoles && !allowedRoles.includes(user.role)) {
      // Redirect based on role
      if (user.role === "student") {
        router.push("/island");
      } else {
        router.push("/overview");
      }
      return;
    }
  }, [isInitialized, accessToken, user, allowedRoles, router, fetchUser, pathname]);

  // Show loading state while checking auth
  if (!isInitialized || !accessToken || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-500"></div>
      </div>
    );
  }

  // Check role access
  if (allowedRoles && user && !allowedRoles.includes(user.role)) {
    return null;
  }

  return <>{children}</>;
}
