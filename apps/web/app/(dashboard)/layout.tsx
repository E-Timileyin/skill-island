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
    <div className="min-h-screen bg-gray-50 font-['Nunito']">
      {/* Navigation */}
      <nav className="border-b border-gray-200 bg-white shadow-sm">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6 lg:px-8">
          <div className="flex items-center gap-6">
            <span className="text-xl font-bold text-indigo-600">
              🏝️ Skill Island
            </span>
            <a
              href="/overview"
              className="text-sm font-semibold text-gray-700 hover:text-indigo-600"
            >
              Dashboard
            </a>
            <span className="cursor-default text-sm text-gray-400">
              Settings
            </span>
          </div>

          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">
              Welcome{user ? `, ${user.email}` : ""}
            </span>
            <button
              onClick={handleLogout}
              className="rounded-lg bg-gray-100 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200"
            >
              Logout
            </button>
          </div>
        </div>
      </nav>

      {/* Content */}
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  );
}
