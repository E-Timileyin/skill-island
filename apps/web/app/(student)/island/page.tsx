"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/hooks/useAuth";
import { getProfile, type Profile } from "@/lib/api";
import IslandMap from "@/components/game/IslandMap";
// import bgMemoryCove from "@/public/assets/bg-memory-cove.png";
// import Image from 'next/image'

/** XP thresholds for the game zones. */
const ZONE_THRESHOLDS: { zone: string; requiredXP: number; deferred?: boolean }[] = [
  { zone: "memory_cove", requiredXP: 0 },
  { zone: "cross_nought", requiredXP: 0 },
  { zone: "focus_forest", requiredXP: 30 },
  { zone: "team_tower", requiredXP: 80 },
  { zone: "pattern_plateau", requiredXP: 150, deferred: true },
  { zone: "community_hub", requiredXP: 250, deferred: true },
];

const XP_SESSION_KEY = "skill_island_prev_xp";

function detectNewlyUnlocked(previousXP: number, currentXP: number): Set<string> {
  const result = new Set<string>();
  for (const z of ZONE_THRESHOLDS) {
    if (z.deferred) continue;
    if (previousXP < z.requiredXP && currentXP >= z.requiredXP) {
      result.add(z.zone);
    }
  }
  return result;
}

export default function IslandPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [profileLoading, setProfileLoading] = useState(true);
  const [newlyUnlocked, setNewlyUnlocked] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (authLoading) return;
    if (!user) {
      router.replace("/login");
      return;
    }

    getProfile()
      .then((p) => {
        // Detect newly unlocked zones by comparing with previous XP
        const prevXPStr = sessionStorage.getItem(XP_SESSION_KEY);
        const parsedXP = prevXPStr !== null ? parseInt(prevXPStr, 10) : NaN;
        const prevXP = Number.isFinite(parsedXP) ? parsedXP : p.total_xp;
        const unlocked = detectNewlyUnlocked(prevXP, p.total_xp);
        setNewlyUnlocked(unlocked);

        // Store current XP for next comparison
        sessionStorage.setItem(XP_SESSION_KEY, String(p.total_xp));

        setProfile(p);
      })
      .catch((err: unknown) => {
        const apiErr = err as { code?: string };
        if (apiErr.code === "PROFILE_NOT_FOUND" || apiErr.code === "NOT_FOUND") {
          router.replace("/setup");
        } else {
          // Network or unexpected error — redirect to setup as fallback
          router.replace("/setup");
        }
      })
      .finally(() => setProfileLoading(false));
  }, [user, authLoading, router]);

  const handleZoneSelect = useCallback(
    (zone: string) => {
      router.push(`/game/${zone}`);
    },
    [router]
  );

  if (authLoading || profileLoading) {
    return (

      <main className="flex min-h-screen flex-col items-center bg-gradient-to-b from-sky-100 to-blue-50 p-4 font-['Nunito']">
        {/* <Image
          src={bgMemoryCove}
          alt="image"
          fill
          priority
          className="absolute"
        /> */}
        {/* Loading skeleton */}
        <div className="w-full max-w-2xl animate-pulse space-y-6 pt-4">
          {/* HUD skeleton */}
          <div className="flex items-center gap-4 rounded-2xl bg-white p-4 shadow-md">
            <div className="h-12 w-12 rounded-full bg-gray-200" />
            <div className="flex flex-col gap-2">
              <div className="h-4 w-24 rounded bg-gray-200" />
              <div className="h-3 w-16 rounded bg-gray-200" />
            </div>
            <div className="ml-auto flex flex-col items-end gap-1 min-w-[140px]">
              <div className="h-3 w-20 rounded bg-gray-200" />
              <div className="h-3 w-full rounded-full bg-gray-200" />
            </div>
          </div>
          {/* Title skeleton */}
          <div className="mx-auto h-8 w-48 rounded bg-gray-200" />
          {/* Cards skeleton */}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div
                key={i}
                className="flex flex-col items-center gap-2 rounded-2xl bg-white p-6 shadow-sm"
              >
                <div className="h-12 w-12 rounded-full bg-gray-200" />
                <div className="h-4 w-24 rounded bg-gray-200" />
                <div className="h-6 w-16 rounded-full bg-gray-200" />
              </div>
            ))}
          </div>
        </div>
      </main>
    );
  }

  if (!profile) return null;

  return (
    <IslandMap
      totalXP={profile.total_xp}
      totalStars={profile.total_stars}
      playerNickname={profile.nickname}
      avatarId={profile.avatar_id}
      onZoneSelect={handleZoneSelect}
      newlyUnlockedZones={newlyUnlocked}
    />
  );
}
