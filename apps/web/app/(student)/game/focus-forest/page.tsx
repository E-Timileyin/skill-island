"use client";

export const dynamic = 'force-dynamic';

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import PhaserGame from "@/game/PhaserGame";
import FocusForestUI from "@/components/game/FocusForestUI";
import eventBus from "@/game/events/EventBus";
import { submitSession } from "@/lib/api";
import SessionResultScreen from "@/components/game/SessionResultScreen";
import Image from "next/image";

export default function FocusForestPage() {
  const router = useRouter();

  const [phase, setPhase] = useState<'playing' | 'complete' | 'saving' | 'error' | 'results'>('playing');
  const [timeRemainingMs, setTimeRemainingMs] = useState(60000);
  const [butterfliesHit, setButterfliesHit] = useState(0);

  const [sessionResult, setSessionResult] = useState<{
    outcome: 'win' | 'lose' | 'incomplete';
    starsEarned: number;
    xpEarned: number;
    unlockedZones: string[];
  } | null>(null);

  useEffect(() => {
    const handleUIUpdate = (data: any) => {
      setTimeRemainingMs(data.timeRemainingMs);
      setButterfliesHit(data.butterfliesHit);
    };

    const handleSessionEnd = async (data: any) => {
      if (data.game_type !== 'focus_forest') return;
      
      setPhase('saving');

      try {
        const result = await submitSession({
          session_token: data.session_token,
          actions: data.actions
        });

        // The session endpoint handles total XP calculation.
        setSessionResult({
          outcome: 'win', // By SEND conventions we default to win presentation.
          starsEarned: result.stars_earned ?? 1,
          xpEarned: result.xp_earned ?? 0,
          unlockedZones: result.unlocked_zones ?? []
        });
        setPhase('results');

      } catch (err) {
        console.error("Failed submitting session:", err);
        setPhase('error');
      }
    };

    eventBus.on("game:ui-update", handleUIUpdate);
    eventBus.on("game:session-end", handleSessionEnd);

    return () => {
      eventBus.off("game:ui-update", handleUIUpdate);
      eventBus.off("game:session-end", handleSessionEnd);
    };
  }, []);

  if (phase === 'results' && sessionResult) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-900 p-8">
        <SessionResultScreen
          outcome={sessionResult.outcome}
          starsEarned={sessionResult.starsEarned}
          xpEarned={sessionResult.xpEarned}
          unlockedZones={sessionResult.unlockedZones}
          onPlayAgain={() => {
            setPhase('playing');
            setTimeRemainingMs(60000);
            setButterfliesHit(0);
            setSessionResult(null);
            // Quick refresh pattern to reboot phaser state perfectly
            window.location.reload();
          }}
          onGoToIsland={() => router.push('/island')}
        />
      </div>
    );
  }

  if (phase === 'error') {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-slate-900 p-8 text-center text-white">
        <h1 className="mb-4 text-3xl font-bold text-rose-400">Something went wrong</h1>
        <p className="mb-8 text-lg text-slate-300">
          Your progress is saved, but we couldn't properly contact the server.
        </p>
        <div className="flex space-x-4">
          <button
            onClick={() => window.location.reload()}
            className="rounded-full bg-blue-600 px-6 py-2 font-semibold hover:bg-blue-500"
          >
            Try Again
          </button>
          <button
            onClick={() => router.push('/island')}
            className="rounded-full bg-slate-700 px-6 py-2 font-semibold hover:bg-slate-600"
          >
            Back to Island
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="relative flex min-h-screen flex-col items-center justify-center p-4 overflow-hidden">
      {/* Background Image */}
      <div className="absolute inset-0 z-0">
        <Image
          src="/assets/images/bg-focus-forest.jpg"
          alt="Focus Forest Background"
          fill
          className="object-cover scale-105"
          priority
        />
        <div className="absolute inset-0 bg-emerald-900/10 backdrop-blur-[1px]" />
      </div>

      <div className="relative z-10 overflow-hidden rounded-[2rem] shadow-[0_20px_50px_rgba(0,0,0,0.5)] border-8 border-white/20 backdrop-blur-sm">
        <FocusForestUI
          timeRemainingMs={timeRemainingMs}
          totalDurationMs={60000}
          butterfliesHit={butterfliesHit}
          phase={phase === 'playing' ? 'playing' : 'complete'}
        />
        {(phase === 'playing' || phase === 'saving' || phase === 'complete') && (
          <PhaserGame scene="FocusForestScene" />
        )}
      </div>
    </div>
  );
}
