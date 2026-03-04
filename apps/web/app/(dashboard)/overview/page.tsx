"use client";

import { useAuth } from "@/hooks/useAuth";

export default function DashboardOverviewPage() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-lg text-gray-500">Loading…</p>
      </main>
    );
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center">
      <h1 className="text-3xl font-bold">📊 Dashboard</h1>
      <p className="mt-2 text-gray-600">
        Welcome{user ? `, ${user.email}` : ""}!
      </p>
      <p className="mt-4 text-lg text-gray-400">Dashboard coming soon</p>
    </main>
  );
}
