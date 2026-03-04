"use client";

import { useEffect, useState } from "react";
import { useAuth } from "@/hooks/useAuth";
import { getAnalyticsOverview, type WeeklySummary } from "@/lib/api";
import ParentDashboard from "@/components/dashboard/ParentDashboard";
import ExportButton from "@/components/dashboard/ExportButton";

export default function DashboardOverviewPage() {
  const { user, loading: authLoading } = useAuth();
  const [summary, setSummary] = useState<WeeklySummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (authLoading || !user) return;

    // For MVP, profile_id comes from query param or a default.
    // In a full implementation, parent-child linking would provide this.
    const params = new URLSearchParams(window.location.search);
    const profileId = params.get("profile_id");

    if (!profileId) {
      setIsLoading(false);
      return;
    }

    getAnalyticsOverview(profileId)
      .then(setSummary)
      .catch(() => {
        // No data or error — summary stays null
      })
      .finally(() => setIsLoading(false));
  }, [user, authLoading]);

  if (authLoading) {
    return (
      <main className="flex min-h-[60vh] items-center justify-center">
        <p className="text-lg text-gray-500">Loading…</p>
      </main>
    );
  }

  const dashboardLabel =
    user?.role === "educator" ? "Educator Dashboard" : "Parent Dashboard";

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">
          📊 {dashboardLabel}
        </h1>
        <ExportButton />
      </div>

      <ParentDashboard
        summary={summary}
        isLoading={isLoading}
        childNickname="Your child"
      />
    </div>
  );
}

