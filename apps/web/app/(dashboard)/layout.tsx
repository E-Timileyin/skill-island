"use client";

import { useAuthStore } from "@/hooks/useAuthStore";
import { useRouter } from "next/navigation";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { logout } = useAuthStore();
  const router = useRouter();

  const handleLogout = async () => {
    logout();
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
