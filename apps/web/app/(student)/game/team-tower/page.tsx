"use client";

export const dynamic = 'force-dynamic';

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import PhaserGame from "@/game/PhaserGame";
import TeamTowerUI from "@/components/game/TeamTowerUI";
import WaitingForPartner from "@/components/game/WaitingForPartner";
import CoopSessionResultScreen from "@/components/game/CoopSessionResultScreen";
import eventBus from "@/game/events/EventBus";
import { submitCoopSession, getProfile, Profile } from "@/lib/api";

type Phase = "init" | "waiting" | "ready" | "playing" | "partner_disconnected" | "idle_warning" | "complete" | "saving" | "results" | "error" | "reconnecting" | "partner_reconnected";

export default function TeamTowerPage() {
  const router = useRouter();

  const [phase, setPhase] = useState<Phase>("init");
  const [profile, setProfile] = useState<Profile | null>(null);
  
  // Game UI States
  const [groupXP, setGroupXP] = useState(0);
  const [activePlayer, setActivePlayer] = useState("");
  const [myRole, setMyRole] = useState("");
  const [turnNumber, setTurnNumber] = useState(1);
  const [opponentAvatarId, setOpponentAvatarId] = useState(0);
  const [idleSeconds, setIdleSeconds] = useState(0);
  const [waitingSeconds, setWaitingSeconds] = useState(0);

  // Result States
  const [sessionResult, setSessionResult] = useState<{
    outcome: 'win' | 'lose' | 'incomplete';
    starsEarned: number;
    groupXPEarned: number;
    myXPEarned: number;
    totalXP: number;
    unlockedZones: string[];
  } | null>(null);

  useEffect(() => {
    // 1. Verify play_mode on load
    getProfile().then(p => {
      setProfile(p);
      if (p.play_mode === "solo") {
        setPhase("error"); // Will show mode mismatch
      } else {
        setPhase("waiting");
      }
    }).catch(() => {
      router.push('/login');
    });
  }, [router]);

  useEffect(() => {
    if (phase === "waiting") {
      const waitInterval = setInterval(() => {
        setWaitingSeconds(s => s + 1);
      }, 1000);
      return () => clearInterval(waitInterval);
    }
  }, [phase]);

  useEffect(() => {
    const handleUIUpdate = (data: any) => {
      if (data.phase) setPhase(data.phase);
      if (data.groupXP !== undefined) setGroupXP(data.groupXP);
      if (data.activePlayer) setActivePlayer(data.activePlayer);
      if (data.myRole) setMyRole(data.myRole);
      if (data.turnNumber !== undefined) setTurnNumber(data.turnNumber);
      if (data.opponentAvatar !== undefined) setOpponentAvatarId(data.opponentAvatar);
      if (data.secondsRemaining !== undefined) setIdleSeconds(data.secondsRemaining);
    };

    const handleSessionEnd = async (data: any) => {
      if (data.game_type !== 'team_tower') return;
      
      setPhase("saving");

      try {
        const result = await submitCoopSession({
          game_type: "team_tower",
          mode: "cooperative",
          room_session_id: data.room_session_id,
          outcome: data.outcome,
          duration_ms: data.duration_ms
        });

        setSessionResult({
          outcome: data.outcome as 'win' | 'lose' | 'incomplete',
          starsEarned: result.stars || data.stars || 1, 
          groupXPEarned: data.group_xp || 0,
          myXPEarned: result.xp_earned || 0,
          totalXP: (profile?.total_xp || 0) + (result.xp_earned || 0), // Use server's exact but estimate here for smooth UX
          unlockedZones: [] 
        });
        
        setPhase("results");
      } catch (err) {
        console.error("Co-op sumbit failed", err);
        setPhase("error");
      }
    };

    eventBus.on("game:ui-update", handleUIUpdate);
    eventBus.on("game:session-end", handleSessionEnd);

    return () => {
      eventBus.off("game:ui-update", handleUIUpdate);
      eventBus.off("game:session-end", handleSessionEnd);
    };
  }, [profile]);


  if (phase === "init") {
    return <div className="min-h-screen bg-slate-900 flex items-center justify-center text-white">Loading Tower...</div>;
  }

  // 2. Profile mode rejection screen
  if (phase === "error" && profile?.play_mode === "solo") {
    return (
      <div className="flex min-h-screen items-center justify-center bg-sky-900 p-8">
        <div className="bg-white rounded-3xl p-10 max-w-lg text-center shadow-2xl">
          <h2 className="text-3xl font-bold text-slate-800 mb-4">You're Playing Solo! 🏝️</h2>
          <p className="text-lg text-slate-600 mb-8">
            Team Tower needs two players! Switch to <strong className="text-sky-600">Team mode</strong> in your profile to play.
          </p>
          <div className="flex justify-center space-x-4">
             <button 
                onClick={() => router.push('/dashboard/profile')} 
                className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-xl font-bold transition-all"
             >
               Update My Profile
             </button>
             <button 
                onClick={() => router.push('/island')} 
                className="bg-slate-200 hover:bg-slate-300 text-slate-700 px-6 py-3 rounded-xl font-bold transition-all"
             >
               Back to Island
             </button>
          </div>
        </div>
      </div>
    );
  }

  // Generic Error handling
  if (phase === "error") {
     return (
        <div className="flex min-h-screen items-center justify-center bg-slate-900">
           <div className="bg-white rounded-xl p-8 text-center max-w-md">
             <h2 className="text-2xl font-bold text-rose-500 mb-2">Something went wrong</h2>
             <p className="text-slate-600 mb-6">Your progress is saved, but we had trouble completing the request.</p>
             <button onClick={() => window.location.reload()} className="bg-blue-500 text-white px-4 py-2 rounded-xl">Try Again</button>
           </div>
        </div>
     );
  }

  if (phase === "results" && sessionResult) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-900">
         <CoopSessionResultScreen
           outcome={sessionResult.outcome}
           starsEarned={sessionResult.starsEarned}
           groupXPEarned={sessionResult.groupXPEarned}
           myXPEarned={sessionResult.myXPEarned}
           totalXP={sessionResult.totalXP}
           unlockedZones={sessionResult.unlockedZones}
           onPlayAgain={() => window.location.reload()}
           onGoToIsland={() => router.push('/island')}
         />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-slate-950 p-4 relative">
      <div className="relative overflow-hidden rounded-2xl shadow-2xl bg-sky-900 border-4 border-slate-800">
        
        {phase === "waiting" && (
          <WaitingForPartner 
             waitingSeconds={waitingSeconds} 
             onCancel={() => router.push('/island')} 
          />
        )}
        
        <TeamTowerUI
          phase={phase}
          groupXP={groupXP}
          groupXPTarget={100}
          activePlayer={activePlayer}
          myRole={myRole}
          turnNumber={turnNumber}
          opponentAvatarId={opponentAvatarId}
          idleSecondsRemaining={idleSeconds}
        />
        
        {(phase !== "waiting" && phase !== "saving" && phase !== "results") && (
           <PhaserGame scene="TeamTowerScene" />
        )}
      </div>
    </div>
  );
}
