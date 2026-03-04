"use client";

interface ZoneCardProps {
  zone: string;
  label: string;
  emoji: string;
  locked: boolean;
  deferred: boolean;
  requiredXP: number;
  currentXP: number;
  isNewlyUnlocked?: boolean;
  onClick: () => void;
}

export default function ZoneCard(props: ZoneCardProps) {
  const {
    label,
    emoji,
    locked,
    deferred,
    requiredXP,
    currentXP,
    isNewlyUnlocked,
    onClick,
  } = props;
  const unlocked = !locked;

  return (
    <button
      type="button"
      onClick={unlocked ? onClick : undefined}
      disabled={locked}
      aria-label={
        deferred
          ? `${label} — Coming Soon`
          : locked
            ? `${label} — Need ${requiredXP - currentXP} XP to unlock`
            : `Play ${label}`
      }
      className={`
        relative flex flex-col items-center justify-center gap-2
        rounded-2xl p-6 w-full
        font-['Nunito'] transition-all duration-300
        ${
          unlocked
            ? `bg-white shadow-lg hover:scale-105 hover:shadow-xl cursor-pointer
               border-2 border-transparent hover:border-blue-300`
            : "bg-gray-100 shadow-sm cursor-not-allowed opacity-60"
        }
        ${isNewlyUnlocked ? "animate-pulse ring-4 ring-yellow-400 ring-offset-2" : ""}
      `}
    >
      {/* Emoji */}
      <span className={`text-5xl ${locked ? "grayscale" : ""}`} aria-hidden="true">
        {locked ? "🔒" : emoji}
      </span>

      {/* Label */}
      <span
        className={`text-lg font-bold ${
          unlocked ? "text-gray-800" : "text-gray-400"
        }`}
      >
        {label}
      </span>

      {/* Status badge */}
      {deferred ? (
        <span className="rounded-full bg-gray-200 px-3 py-1 text-xs font-semibold text-gray-500">
          Coming Soon
        </span>
      ) : locked ? (
        <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-700">
          Need {requiredXP - currentXP} XP
        </span>
      ) : (
        <span className="rounded-full bg-green-100 px-3 py-1 text-sm font-semibold text-green-700">
          Play
        </span>
      )}
    </button>
  );
}
