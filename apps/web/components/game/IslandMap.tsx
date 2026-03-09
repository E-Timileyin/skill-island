"use client";

import PlayerHUD from "./PlayerHUD";
import Image from "next/image";
import { Lock } from "lucide-react";
import { motion } from "motion/react";

interface ZoneData {
  zone: string;
  label: string;
  requiredXP: number;
  deferred?: boolean;
  top: string;
  left: string;
}

const ZONES: ZoneData[] = [
  { zone: "memory_cove", label: "Memory Cove", requiredXP: 0, top: "32%", left: "15%" },
  { zone: "cross_nought", label: "Cross & Nought", requiredXP: 0, top: "55%", left: "18%" },
  { zone: "pattern_plateau", label: "Pattern Plateau", requiredXP: 150, top: "25%", left: "50%", deferred: true },
  { zone: "team_tower", label: "Team Tower", requiredXP: 80, top: "28%", left: "82%" },
  { zone: "focus_forest", label: "Focus Forest", requiredXP: 30, top: "52%", left: "78%" },
  { zone: "community_hub", label: "Community Hub", requiredXP: 250, deferred: true, top: "75%", left: "50%" },
];

interface IslandMapProps {
  totalXP: number;
  totalStars: number;
  playerNickname: string;
  avatarId: number;
  onZoneSelect: (zone: string) => void;
  newlyUnlockedZones?: Set<string>;
}

export default function IslandMap({
  totalXP,
  totalStars,
  playerNickname,
  avatarId,
  onZoneSelect,
  newlyUnlockedZones,
}: IslandMapProps) {
  return (
    <div className="relative min-h-screen w-full flex flex-col font-['Nunito'] overflow-hidden">
      {/* Main Full-Screen Map Background */}
      <div className="absolute inset-0 z-0">
        <Image
          src="/assets/images/bg-island.jpeg"
          alt="Island Map"
          fill
          className="object-cover"
          priority
        />
      </div>

      {/* Top Header Section */}
      <div className="relative z-20 flex flex-col items-center w-full max-w-7xl mx-auto mb-4 p-4 sm:p-6 lg:p-8">
        <div className="w-full flex justify-between items-start mb-2">
          <PlayerHUD
            nickname={playerNickname}
            avatarId={avatarId}
            totalStars={totalStars}
            totalXP={totalXP}
          />
        </div>
        
        {/* Hanging Skill Island Sign */}
        <div className="relative hidden md:flex flex-col items-center -mt-8 mb-4">
          <div className="flex gap-24 h-8">
            <div className="w-2 h-full border-l-2 border-r-2 border-amber-800 bg-amber-600 rounded-sm shadow-[inset_0_0_4px_rgba(0,0,0,0.5)] bg-[repeating-linear-gradient(45deg,transparent,transparent_2px,rgba(0,0,0,0.1)_2px,rgba(0,0,0,0.1)_4px)]" />
            <div className="w-2 h-full border-l-2 border-r-2 border-amber-800 bg-amber-600 rounded-sm shadow-[inset_0_0_4px_rgba(0,0,0,0.5)] bg-[repeating-linear-gradient(45deg,transparent,transparent_2px,rgba(0,0,0,0.1)_2px,rgba(0,0,0,0.1)_4px)]" />
          </div>
          <div className="bg-[#cd8347] border-4 border-[#7a4b27] rounded-xl px-12 py-3 shadow-[0_6px_0_#7a4b27] transform -skew-x-2">
            <h1 className="text-4xl md:text-5xl font-extrabold text-white tracking-widest uppercase drop-shadow-[0_2px_2px_rgba(0,0,0,0.6)]" style={{ textShadow: "2px 2px 0 #7a4b27, -2px -2px 0 #7a4b27, 2px -2px 0 #7a4b27, -2px 2px 0 #7a4b27, 0 4px 6px rgba(0,0,0,0.5)"}}>
              Skill Island
            </h1>
          </div>
        </div>
      </div>

      {/* Zones Layer */}
      <div className="absolute inset-0 z-10 pointer-events-none">
        <div className="relative w-full h-full max-w-7xl mx-auto">
          {ZONES.map((z, index) => {
            const isLocked = totalXP < z.requiredXP || z.deferred === true;
            const isNewlyUnlocked = newlyUnlockedZones?.has(z.zone);

            return (
              <motion.button
                key={z.zone}
                onClick={!isLocked ? () => onZoneSelect(z.zone.replace('_', '-')) : undefined}
                disabled={isLocked}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 + 0.3, type: "spring", stiffness: 200 }}
                whileHover={!isLocked ? { scale: 1.1, rotate: [-1, 1, -1, 0] } : undefined}
                whileTap={!isLocked ? { scale: 0.95 } : undefined}
                className={`absolute transform -translate-x-1/2 -translate-y-1/2 pointer-events-auto flex flex-col items-center group`}
                style={{ top: z.top, left: z.left }}
              >
                {isLocked ? (
                  /* Locked State */
                  <div className="bg-slate-900/60 backdrop-blur-sm rounded-2xl py-3 px-5 flex flex-col items-center justify-center gap-2 border-2 border-white/20 shadow-xl transition-all hover:bg-slate-900/70">
                    <span className="text-white/90 text-sm md:text-lg font-bold tracking-wide whitespace-nowrap">
                      {z.label}
                    </span>
                    <Lock size={32} className="text-white/90 drop-shadow-md" strokeWidth={2.5} />
                    {z.deferred ? (
                      <span className="text-xs text-white/70 font-bold tracking-wider uppercase">LOCKED</span>
                    ) : (
                      <span className="text-xs text-amber-300 font-bold">Need {z.requiredXP} XP</span>
                    )}
                  </div>
                ) : (
                  /* Unlocked State (Wooden Sign) */
                  <div className={`relative ${isNewlyUnlocked ? "animate-bounce" : ""}`}>
                    {isNewlyUnlocked && (
                      <div className="absolute -inset-4 bg-yellow-400/30 blur-xl rounded-full z-0 pointer-events-none" />
                    )}
                    <div className="relative z-10 bg-[#cd8347] border-[3px] md:border-4 border-[#7a4b27] rounded-lg px-4 py-2 shadow-[0_6px_0_#7a4b27,0_10px_10px_rgba(0,0,0,0.3)] group-hover:bg-[#d69055] transition-colors">
                      <span className="text-white text-sm md:text-lg font-extrabold tracking-wide whitespace-nowrap drop-shadow-[0_2px_1px_rgba(0,0,0,0.5)]">
                        {z.label}
                      </span>
                    </div>
                  </div>
                )}
              </motion.button>
            );
          })}
        </div>
      </div>
    </div>
  );
}
