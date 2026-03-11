"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/hooks/useAuthStore";
import { setAuthCallbacks } from "@/lib/api";
import { createProfile, getProfile } from "@/lib/api";
import ProfileSetup from "@/components/game/ProfileSetup";

export default function SetupPage() {
  const router = useRouter();
  const { user, accessToken, fetchUser } = useAuthStore();
  const [checkingProfile, setCheckingProfile] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | undefined>();

  // Initialize auth callbacks
  useEffect(() => {
    setAuthCallbacks(
      () => useAuthStore.getState().accessToken,
      () => useAuthStore.getState().refreshTokens()
    );
  }, []);

  useEffect(() => {
    // No access token - redirect to login
    if (!accessToken) {
      router.replace("/login");
      return;
    }

    // Fetch user if we have token but no user
    if (accessToken && !user) {
      fetchUser().catch(() => router.replace("/login"));
      return;
    }

    if (!user) return;

    // Check if profile already exists → stay on setup if not
    getProfile()
      .then(() => {
        router.replace("/island");
      })
      .catch(() => {
        // No profile yet — stay on setup
        setCheckingProfile(false);
      });
  }, [user, accessToken, router, fetchUser]);

  async function handleSubmit(data: {
    nickname: string;
    avatar_id: number;
    play_mode: string;
  }) {
    setError(undefined);
    setSubmitting(true);

    try {
      await createProfile(data);
      router.push("/island");
    } catch (err: unknown) {
      const apiErr = err as { message?: string };
      setError(apiErr.message || "Failed to create profile");
    } finally {
      setSubmitting(false);
    }
  }

  if (!accessToken || !user || checkingProfile) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-lg text-gray-500">Loading…</p>
      </main>
    );
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-gradient-to-b from-sky-100 to-blue-50 p-4">
      <ProfileSetup
        onSubmit={handleSubmit}
        isLoading={submitting}
        error={error}
      />
    </main>
  );
}
