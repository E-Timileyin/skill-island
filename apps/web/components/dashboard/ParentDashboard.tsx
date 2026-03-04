"use client";

import type { WeeklySummary } from "@/lib/api";

interface ParentDashboardProps {
  summary: WeeklySummary | null;
  isLoading: boolean;
  childNickname: string;
}

function getEncouragingLabel(score: number | null): string {
  if (score === null) return "No data yet";
  if (score >= 80) return "Great Concentration!";
  if (score >= 50) return "Keep Going!";
  return "Getting Started!";
}

function getMemoryLabel(score: number | null): string {
  if (score === null) return "No data yet";
  if (score >= 80) return "Excellent Memory!";
  if (score >= 50) return "Good Progress!";
  return "Keep Practising!";
}

function formatPercent(value: number | null): string {
  if (value === null) return "—";
  return `${Math.round(value * 100)}%`;
}

function SkeletonCard() {
  return (
    <div className="animate-pulse rounded-2xl bg-gray-100 p-6 h-44">
      <div className="h-8 w-8 rounded-full bg-gray-200 mb-3" />
      <div className="h-6 w-20 rounded bg-gray-200 mb-2" />
      <div className="h-4 w-32 rounded bg-gray-200" />
    </div>
  );
}

export default function ParentDashboard({
  summary,
  isLoading,
  childNickname,
}: ParentDashboardProps) {
  if (isLoading) {
    return (
      <div className="space-y-6">
        <h2 className="text-2xl font-bold text-gray-800">
          Loading {childNickname}&apos;s progress…
        </h2>
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      </div>
    );
  }

  const noData = !summary || summary.message === "No data yet";

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold text-gray-800">
        {childNickname}&apos;s Weekly Summary
      </h2>

      {noData && (
        <p className="text-gray-500">
          No analytics data available yet. Data will appear after game sessions
          are played.
        </p>
      )}

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {/* Attention Card */}
        <div className="rounded-2xl bg-blue-50 p-6 shadow-sm">
          <div className="mb-2 text-3xl">🦉</div>
          <p className="text-sm font-semibold uppercase tracking-wide text-blue-600">
            Attention
          </p>
          <p className="mt-1 text-3xl font-bold text-blue-900">
            {noData ? "—" : formatPercent(summary?.attention_score ?? null)}
          </p>
          <p className="mt-1 text-sm text-blue-700">
            {getEncouragingLabel(summary?.attention_score ?? null)}
          </p>
        </div>

        {/* Memory Card */}
        <div className="rounded-2xl bg-purple-50 p-6 shadow-sm">
          <div className="mb-2 text-3xl">🐘</div>
          <p className="text-sm font-semibold uppercase tracking-wide text-purple-600">
            Memory
          </p>
          <p className="mt-1 text-3xl font-bold text-purple-900">
            {noData ? "—" : formatPercent(summary?.memory_score ?? null)}
          </p>
          <p className="mt-1 text-sm text-purple-700">
            {getMemoryLabel(summary?.memory_score ?? null)}
          </p>
        </div>

        {/* Social Engagement Card */}
        <div className="rounded-2xl bg-green-50 p-6 shadow-sm">
          <div className="mb-2 text-3xl">🤝</div>
          <p className="text-sm font-semibold uppercase tracking-wide text-green-600">
            Social Engagement
          </p>
          <p className="mt-1 text-3xl font-bold text-green-900">
            {noData
              ? "—"
              : formatPercent(summary?.coop_participation_rate ?? null)}
          </p>
          <p className="mt-1 text-sm text-green-700">
            {noData
              ? "No data yet"
              : `${summary?.sessions_this_week ?? 0} sessions this week`}
          </p>
        </div>

        {/* Progress Card */}
        <div className="rounded-2xl bg-amber-50 p-6 shadow-sm">
          <div className="mb-2 text-3xl">⭐</div>
          <p className="text-sm font-semibold uppercase tracking-wide text-amber-600">
            Progress
          </p>
          <p className="mt-1 text-3xl font-bold text-amber-900">
            {noData ? "0" : (summary?.total_xp ?? 0)} XP
          </p>
          <p className="mt-1 text-sm text-amber-700">
            {noData
              ? "No data yet"
              : `${summary?.total_stars ?? 0} stars · ${summary?.sessions_this_week ?? 0} sessions`}
          </p>
        </div>
      </div>
    </div>
  );
}
