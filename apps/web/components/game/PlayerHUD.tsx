"use client";

import { AVATARS } from "@/lib/avatars";
import Image from "next/image";

interface PlayerHUDProps {
  nickname: string;
  avatarId: number;
  totalStars: number;
  totalXP: number;
}

/** XP thresholds for zone unlocks, in ascending order. */
const XP_THRESHOLDS = [0, 30, 80, 150, 250];

function getNextThreshold(xp: number): number {
  for (const t of XP_THRESHOLDS) {
    if (xp < t) return t;
  }
  return XP_THRESHOLDS[XP_THRESHOLDS.length - 1];
}

function getPreviousThreshold(xp: number): number {
  let prev = 0;
  for (const t of XP_THRESHOLDS) {
    if (xp < t) return prev;
    prev = t;
  }
  return prev;
}

export default function PlayerHUD({
  nickname,
  avatarId,
  totalStars,
  totalXP,
}: PlayerHUDProps) {
  const avatarObj = AVATARS.find((a) => a.id === avatarId) || AVATARS[0];
  const nextThreshold = getNextThreshold(totalXP);
  const prevThreshold = getPreviousThreshold(totalXP);
  const range = nextThreshold - prevThreshold;
  const progress = range > 0 ? ((totalXP - prevThreshold) / range) * 100 : 100;

  return (
    <div className="flex w-full items-center gap-4 rounded-2xl bg-white p-4 shadow-md font-['Nunito']">
      {/* Avatar */}
      <div className="relative flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-blue-100 overflow-hidden border-2 border-blue-200">
        <Image
          src={avatarObj.src}
          alt={nickname}
          fill
          className="object-cover"
        />
      </div>

      {/* Name + stars */}
      <div className="flex flex-col">
        <span className="text-lg font-bold text-gray-800">{nickname}</span>
        <span className="text-sm text-yellow-600 font-semibold">
          ⭐ {totalStars} stars
        </span>
      </div>

      {/* XP progress bar */}
      <div className="ml-auto flex flex-col items-end gap-1 min-w-[140px]">
        <span className="text-xs font-semibold text-gray-500">
          {totalXP} / {nextThreshold} XP
        </span>
        <div className="h-3 w-full overflow-hidden rounded-full bg-gray-200">
          <div
            className="h-full rounded-full bg-gradient-to-r from-blue-400 to-blue-600 transition-all duration-700 ease-out"
            style={{ width: `${Math.min(progress, 100)}%` }}
          />
        </div>
      </div>
    </div>
  );
}
