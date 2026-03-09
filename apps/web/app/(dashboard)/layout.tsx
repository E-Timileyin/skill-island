"use client";

import { useAuth } from "@/hooks/useAuth";
import { logout } from "@/lib/api";
import { useRouter } from "next/navigation";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user } = useAuth();
  const router = useRouter();

  const handleLogout = async () => {
    try {
      await logout();
    } catch {
      // Clear cookies even if API call fails
    }
    router.push("/login");
  };

  return (
    <div className="min-h-screen font-['Nunito']">
      <main className="w-full">
        {children}
      </main>
    </div>
  );
}
