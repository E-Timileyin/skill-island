"use client";

import ZoneCard from "./ZoneCard";
import PlayerHUD from "./PlayerHUD";

interface ZoneData {
  zone: string;
  label: string;
  requiredXP: number;
  emoji: string;
  deferred?: boolean;
}

const ZONES: ZoneData[] = [
  { zone: "memory_cove", label: "Memory Cove", requiredXP: 0, emoji: "🧠" },
  { zone: "focus_forest", label: "Focus Forest", requiredXP: 30, emoji: "🦋" },
  { zone: "team_tower", label: "Team Tower", requiredXP: 80, emoji: "🏗️" },
  {
    zone: "pattern_plateau",
    label: "Pattern Plateau",
    requiredXP: 150,
    emoji: "🔢",
    deferred: true,
  },
  {
    zone: "community_hub",
    label: "Community Hub",
    requiredXP: 250,
    emoji: "🏘️",
    deferred: true,
  },
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
    <div className="flex min-h-screen flex-col items-center bg-gradient-to-b from-sky-100 to-blue-50 p-4 font-['Nunito']">
      {/* Player HUD */}
      <div className="w-full max-w-2xl mb-6">
        <PlayerHUD
          nickname={playerNickname}
          avatarId={avatarId}
          totalStars={totalStars}
          totalXP={totalXP}
        />
      </div>

      {/* Island title */}
      <h1 className="mb-6 text-3xl font-bold text-gray-800">
        🏝️ Skill Island
      </h1>

      {/* Zone cards grid */}
      <div className="grid w-full max-w-2xl grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {ZONES.map((z) => {
          const locked = totalXP < z.requiredXP || z.deferred === true;
          return (
            <ZoneCard
              key={z.zone}
              zone={z.zone}
              label={z.label}
              emoji={z.emoji}
              locked={locked}
              deferred={z.deferred === true}
              requiredXP={z.requiredXP}
              currentXP={totalXP}
              isNewlyUnlocked={newlyUnlockedZones?.has(z.zone)}
              onClick={() => onZoneSelect(z.zone)}
            />
          );
        })}
      </div>
    </div>
  );
}
