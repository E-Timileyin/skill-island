"use client";

import dynamic from 'next/dynamic';
import { useEffect, useState } from 'react';
import EventBus from '@/game/events/EventBus';
import { MemoryCoveScreen } from '@/components/game/MemoryCoveScreen';
import SessionResultScreen from '@/components/game/SessionResultScreen';
import { submitSession } from '@/lib/api';
import Image from 'next/image';

const PhaserGame = dynamic(() => import('@/game/PhaserGame'), { ssr: false });

export default function MemoryCovePage() {
  const [phase, setPhase] = useState<'watching' | 'your_turn' | 'complete'>('watching');
  const [currentRound, setCurrentRound] = useState(1);
  const [totalRounds, setTotalRounds] = useState(10);
  const [stars, setStars] = useState(0);
  const [playerNickname, setPlayerNickname] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    const uiUpdate = ({ round, stars, phase }: any) => {
      setCurrentRound(round);
      setStars(stars);
      setPhase(phase);
    };
    EventBus.on('game:ui-update', uiUpdate);
    const profileLoaded = ({ nickname }: any) => setPlayerNickname(nickname);
    const sessionEnd = async ({ sessionToken, actions }: any) => {
      setLoading(true);
      setError('');
      try {
        const res = await submitSession({ session_token: sessionToken, actions });
        setResult(res);
        setPhase('complete');
      } catch (e) {
        setError('Something went wrong');
      } finally {
        setLoading(false);
      }
    };

    EventBus.on('game:ui-update', uiUpdate);
    EventBus.on('game:profile-loaded', profileLoaded);
    EventBus.on('game:session-end', sessionEnd);
    return () => {
      EventBus.off('game:ui-update', uiUpdate);
      EventBus.off('game:profile-loaded', profileLoaded);
      EventBus.off('game:session-end', sessionEnd);
    };
  }, []);

  if (loading) {
    return <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-60 z-50"><span className="text-white text-2xl">Submitting...</span></div>;
  }

  if (result) {
    return <SessionResultScreen {...result} onPlayAgain={() => window.location.reload()} onGoToIsland={() => window.location.href = '/student/island'} />;
  }

  if (error) {
    return <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-60 z-50">
      <div className="bg-white rounded-lg p-6 shadow-lg">
        <p className="text-xl text-red-600 mb-4">{error}</p>
        <button className="btn" onClick={() => window.location.reload()}>Retry</button>
      </div>
    </div>;
  }

  return (
    <main className="relative flex flex-col items-center justify-center min-h-screen p-4 overflow-hidden">
      {/* Background Image */}
      <div className="absolute inset-0 z-0">
        <Image
          src="/assets/images/bg-memory-cove.jpg"
          alt="Memory Cove Background"
          fill
          className="object-cover"
          priority
        />
        <div className="absolute inset-0 bg-blue-900/10 backdrop-blur-[1px]" />
      </div>

      <div className="relative z-10 w-full max-w-4xl flex flex-col items-center">
        <div className="w-full relative overflow-hidden rounded-[2.5rem] shadow-[0_25px_60px_rgba(0,0,0,0.6)] border-8 border-white/30 backdrop-blur-sm">
          <MemoryCoveScreen />
          <PhaserGame scene="MemoryCoveScene" />
        </div>
      </div>
    </main>
  );
}
